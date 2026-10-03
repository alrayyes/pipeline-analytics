import { redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';

// The flaky view supersedes the Steps page: it lists the same flaky steps,
// with their run history. A bookmarked /steps lands there.
export const load: PageLoad = () => {
	redirect(307, '/flaky');
};
