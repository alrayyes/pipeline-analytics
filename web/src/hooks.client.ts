import type { HandleClientError } from '@sveltejs/kit';
import { shapeError } from '#lib/errors.js';

// The app renders client-side (ssr = false), so this is where an error from
// a load function or a component reaches the user. See errors.ts.
export const handleError: HandleClientError = shapeError;
