// Writes static/releases.json from the repository's CHANGELOG.md, so the
// release history page ships with the build instead of querying GitHub on
// every visit (#368). Runs as `prebuild` and `predev`; the output is
// generated, so it's gitignored rather than committed beside the changelog
// it has to agree with.
//
// The release being cut is already in CHANGELOG.md when the release workflow
// builds: release-please writes its section in the release PR, which merges
// before the tag, and .goreleaser.yml's `before` hook builds from that commit.

import { mkdirSync, readFileSync, writeFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { parseChangelog } from '../src/lib/changelog.ts';

const REPO_URL = 'https://github.com/alrayyes/pipeline-analytics';

const root = join(import.meta.dir, '..', '..');
const output = join(import.meta.dir, '..', 'static', 'releases.json');

const releases = parseChangelog(
	readFileSync(join(root, 'CHANGELOG.md'), 'utf8'),
	REPO_URL,
);

mkdirSync(dirname(output), { recursive: true });
writeFileSync(output, `${JSON.stringify(releases, null, '\t')}\n`);

console.log(`releases.json: ${releases.length} releases`);
