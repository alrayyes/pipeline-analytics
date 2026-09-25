// Registers the dashboard's read views as WebMCP tools
// (https://github.com/webmachinelearning/webmcp) for an in-browser agent
// sharing the page with its user (#296). Mirrors serviceWorker.ts's shape:
// a feature check, then a fire-and-forget registration that's a no-op on
// any browser that doesn't support the API.
//
// Every tool's execute() calls the exact same function the corresponding
// page uses to fetch its data, rather than a second implementation of the
// same request -- see design.md's "second path to the same API" decision.

import {
	fetchPipeline,
	fetchPipelines,
	fetchRepoUsage,
} from './dashboardApi.js';

type ModelContext = NonNullable<Document['modelContext']>;

function textResult(data: unknown) {
	return { content: [{ type: 'text' as const, text: JSON.stringify(data) }] };
}

export function registerWebMCPTools(
	modelContext: ModelContext | undefined = typeof document === 'undefined'
		? undefined
		: document.modelContext,
): void {
	if (!modelContext) return;

	modelContext.registerTool({
		name: 'list_pipelines',
		description:
			'List tracked pipelines, grouped by repository, with their current health status.',
		execute: async (input) => {
			const data = await fetchPipelines({
				limit: typeof input.limit === 'number' ? input.limit : 20,
				offset: typeof input.offset === 'number' ? input.offset : 0,
				forge: input.forge as 'github' | 'forgejo' | undefined,
				repoId: input.repoId as string | undefined,
			});

			return textResult(data);
		},
	});

	modelContext.registerTool({
		name: 'get_pipeline',
		description:
			"Get a pipeline's health status and its duration and failure-rate trends.",
		execute: async (input) => {
			const data = await fetchPipeline(input.pipelineId as string);

			return textResult(data);
		},
	});

	modelContext.registerTool({
		name: 'get_repo_usage',
		description:
			"Get a repository's runner-minutes usage, broken down by workflow.",
		execute: async (input) => {
			const data = await fetchRepoUsage(input.repoId as string);

			return textResult(data);
		},
	});
}
