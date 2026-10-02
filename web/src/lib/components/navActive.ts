// Overview is the root, so a prefix match would light it up everywhere.
export function isActivePath(pathname: string, href: string): boolean {
	if (href === '/') return pathname === '/';

	return pathname.startsWith(href);
}
