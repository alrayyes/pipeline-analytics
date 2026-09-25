// Ambient declaration for the WebMCP `document.modelContext` API
// (https://github.com/webmachinelearning/webmcp), scoped to the shape
// `webmcpTools.ts` actually calls -- not the full draft spec surface.
// TypeScript's bundled DOM types don't know about this API yet (it isn't in
// any released lib.dom.d.ts), and it's a pre-stable, actively-changing
// draft, so this is deliberately narrow and kept out of `app.d.ts`
// (SvelteKit's own generated-shape file, not the place for a third-party/
// experimental browser API's types).

export {};

declare global {
	interface Document {
		modelContext?: {
			registerTool: (
				tool: {
					name: string;
					description: string;
					title?: string;
					inputSchema?: unknown;
					annotations?: Record<string, unknown>;
					execute: (
						input: Record<string, unknown>,
						context: { signal: AbortSignal },
					) => Promise<{ content: { type: 'text'; text: string }[] }>;
				},
				options?: { signal?: AbortSignal },
			) => void;
		};
	}
}
