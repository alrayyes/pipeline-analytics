// Turns release-please's CHANGELOG.md into the release list the history page
// shows, so the page can ship with the build instead of asking GitHub for
// the same text on every visit (#368). Pure, so bun test covers it like the
// rest of #lib; scripts/generate-releases.ts is the only caller that touches
// the filesystem.

export interface Release {
	version: string;
	tag: string;
	name: string;
	url: string;
	// YYYY-MM-DD: all the changelog records, and all the page shows.
	date: string;
	// The section's markdown without its "## [version](compare) (date)"
	// heading, which the card header already shows.
	body: string;
}

// "## [0.54.0](https://.../compare/v0.53.0...v0.54.0) (2026-10-02)". The
// bracketed, linked form is what release-please writes; a plain
// "## 0.1.0 (2026-09-01)" is accepted too.
const RELEASE_HEADING =
	/^##\s+\[?(\d+\.\d+\.\d+[^\]\s)]*)\]?(?:\([^)]*\))?\s+\((\d{4}-\d{2}-\d{2})\)\s*$/;

// Level-2 only: a release's own "### Features" sections are body, not
// boundaries.
const SECTION_HEADING = /^##\s/;

// A changelog with no recognisable release heading is an empty list, not an
// error: a format change should read as "No releases yet", not break the
// build.
export function parseChangelog(markdown: string, repoUrl: string): Release[] {
	const releases: Release[] = [];
	let current: { release: Release; lines: string[] } | null = null;

	const finish = () => {
		if (current) current.release.body = current.lines.join('\n').trim();
		current = null;
	};

	for (const line of markdown.replaceAll('\r\n', '\n').split('\n')) {
		const heading = RELEASE_HEADING.exec(line);

		if (heading) {
			finish();

			const version = heading[1];
			const release: Release = {
				version,
				tag: `v${version}`,
				name: `v${version}`,
				url: `${repoUrl}/releases/tag/v${version}`,
				date: heading[2],
				body: '',
			};
			releases.push(release);
			current = { release, lines: [] };
		} else if (SECTION_HEADING.test(line)) {
			// Some other level-2 heading ("## Unreleased"): it ends the previous
			// release and its content belongs to no release.
			finish();
		} else {
			current?.lines.push(line);
		}
	}

	finish();

	return releases;
}
