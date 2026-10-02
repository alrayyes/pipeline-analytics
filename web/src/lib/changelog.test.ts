import { describe, expect, test } from 'bun:test';
import { existsSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import { parseChangelog } from './changelog.js';

const REPO = 'https://github.com/alrayyes/pipeline-analytics';

const TWO_RELEASES = `# Changelog

## [0.54.0](${REPO}/compare/v0.53.0...v0.54.0) (2026-10-02)


### Features

* **api:** serve the failure insights ([#365](${REPO}/issues/365))

### Bug Fixes

* **web:** a fix

## [0.53.0](${REPO}/compare/v0.52.0...v0.53.0) (2026-09-30)


### Features

* **metrics:** older feature
`;

describe('parseChangelog', () => {
	test('reads the version, tag, name, url and date from a release heading', () => {
		const [release] = parseChangelog(TWO_RELEASES, REPO);

		expect(release.version).toBe('0.54.0');
		expect(release.tag).toBe('v0.54.0');
		expect(release.name).toBe('v0.54.0');
		expect(release.url).toBe(`${REPO}/releases/tag/v0.54.0`);
		expect(release.date).toBe('2026-10-02');
	});

	test('keeps the file order, which is newest first', () => {
		expect(parseChangelog(TWO_RELEASES, REPO).map((r) => r.version)).toEqual([
			'0.54.0',
			'0.53.0',
		]);
	});

	test('the body is the section without its own heading, trimmed', () => {
		const [release] = parseChangelog(TWO_RELEASES, REPO);

		expect(release.body.startsWith('### Features')).toBe(true);
		expect(release.body).not.toContain('## [0.54.0]');
		expect(release.body).toBe(release.body.trim());
	});

	test('a release body keeps its own ### sections', () => {
		const [release] = parseChangelog(TWO_RELEASES, REPO);

		expect(release.body).toContain('### Features');
		expect(release.body).toContain('### Bug Fixes');
		expect(release.body).toContain('a fix');
	});

	test('a release body stops at the next release and does not leak it', () => {
		const [newest] = parseChangelog(TWO_RELEASES, REPO);

		expect(newest.body).not.toContain('older feature');
	});

	test('the last release runs to the end of the file', () => {
		const releases = parseChangelog(TWO_RELEASES, REPO);

		expect(releases[1].body).toContain('older feature');
	});

	test('a changelog with no release headings is an empty list, not an error', () => {
		expect(parseChangelog('# Changelog\n\nNothing yet.\n', REPO)).toEqual([]);
		expect(parseChangelog('', REPO)).toEqual([]);
	});

	test('a release with no body has an empty one', () => {
		const [release] = parseChangelog(
			`## [0.1.0](${REPO}/compare/a...b) (2026-09-01)\n\n## [0.0.9](${REPO}/compare/a...b) (2026-08-30)\n\n### Features\n\n* x\n`,
			REPO,
		);

		expect(release.version).toBe('0.1.0');
		expect(release.body).toBe('');
	});

	test('a heading without a compare link is still a release', () => {
		const [release] = parseChangelog(
			'## 0.1.0 (2026-09-01)\n\n### Features\n\n* first\n',
			REPO,
		);

		expect(release.version).toBe('0.1.0');
		expect(release.date).toBe('2026-09-01');
	});

	test('a heading that is not a release ends the previous section and is skipped', () => {
		const releases = parseChangelog(
			`## [0.2.0](${REPO}/compare/a...b) (2026-09-02)\n\n### Features\n\n* kept\n\n## Unreleased\n\n* not a release\n`,
			REPO,
		);

		expect(releases).toHaveLength(1);
		expect(releases[0].body).toContain('kept');
		expect(releases[0].body).not.toContain('not a release');
	});

	test('Windows line endings parse the same', () => {
		const releases = parseChangelog(
			TWO_RELEASES.replaceAll('\n', '\r\n'),
			REPO,
		);

		expect(releases.map((r) => r.version)).toEqual(['0.54.0', '0.53.0']);
		expect(releases[0].body).not.toContain('\r');
	});

	test('a pre-release suffix stays part of the version', () => {
		const [release] = parseChangelog(
			`## [1.0.0-rc.1](${REPO}/compare/a...b) (2026-09-02)\n\n* x\n`,
			REPO,
		);

		expect(release.version).toBe('1.0.0-rc.1');
		expect(release.tag).toBe('v1.0.0-rc.1');
	});
});

const CHANGELOG_PATH = join(import.meta.dir, '..', '..', '..', 'CHANGELOG.md');

// Guards the format assumption against release-please changing it: every
// "## [" heading in the real file must come out as a release. Skipped where
// the file isn't reachable from the test's own location: Stryker runs the
// suite from a copy of web/, which has no repository root above it.
describe.skipIf(!existsSync(CHANGELOG_PATH))(
	"this repository's own CHANGELOG.md",
	() => {
		const real = existsSync(CHANGELOG_PATH)
			? readFileSync(CHANGELOG_PATH, 'utf8')
			: '';

		test('every release heading parses', () => {
			const headings = real
				.split('\n')
				.filter((line) => line.startsWith('## ['));

			expect(parseChangelog(real, REPO)).toHaveLength(headings.length);
		});

		test('every release has a version, a date and a non-empty body', () => {
			for (const release of parseChangelog(real, REPO)) {
				expect(release.version).toMatch(/^\d+\.\d+\.\d+/);
				expect(release.date).toMatch(/^\d{4}-\d{2}-\d{2}$/);
				expect(release.body.length).toBeGreaterThan(0);
			}
		});
	},
);
