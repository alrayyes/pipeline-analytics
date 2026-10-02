import { redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';

// The Pipelines list moved to /pipelines so `/` can be the failure overview
// (#335). Until that lands, `/` sends visitors, and old bookmarks, there.
export const load: PageLoad = () => {
	redirect(307, '/pipelines');
};
