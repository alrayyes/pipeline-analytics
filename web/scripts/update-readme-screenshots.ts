// Rewrites the README's screenshot section from screenshot-routes.ts. Run by
// capture-screenshots.sh after the images are captured, so the release job's
// pull request carries the images and the README entries together.
import { fileURLToPath } from 'node:url';
import { replaceScreenshotSection } from './screenshot-routes.js';

const README = fileURLToPath(new URL('../../README.md', import.meta.url));

const current = await Bun.file(README).text();
await Bun.write(README, replaceScreenshotSection(current));
