import { describe, expect, test } from 'bun:test';
import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { routePatterns } from './generate-routes.ts';

function tree(files: string[]): string {
	const dir = mkdtempSync(join(tmpdir(), 'routes-'));
	for (const file of files) {
		mkdirSync(join(dir, file, '..'), { recursive: true });
		writeFileSync(join(dir, file), '');
	}
	return dir;
}

describe('routePatterns', () => {
	test('lists a route for every directory that holds a page', () => {
		const dir = tree([
			'+page.svelte',
			'runs/+page.svelte',
			'runs/[id]/+page.svelte',
			'pipelines/[id]/flaky-runs/+page.svelte',
		]);
		try {
			expect(routePatterns(dir)).toEqual([
				'/',
				'/pipelines/[id]/flaky-runs',
				'/runs',
				'/runs/[id]',
			]);
		} finally {
			rmSync(dir, { recursive: true });
		}
	});

	test('counts a page that is only a load function, like a redirect', () => {
		const dir = tree(['steps/+page.ts']);
		try {
			expect(routePatterns(dir)).toEqual(['/steps']);
		} finally {
			rmSync(dir, { recursive: true });
		}
	});

	test('skips layouts and error views, which are not routes', () => {
		const dir = tree([
			'+layout.svelte',
			'+layout.ts',
			'+error.svelte',
			'legal/+page.svelte',
		]);
		try {
			expect(routePatterns(dir)).toEqual(['/legal']);
		} finally {
			rmSync(dir, { recursive: true });
		}
	});
});
