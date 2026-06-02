/**
 * Pure URL helpers for the drawer — no DOM side effects, fully unit-testable.
 */

export interface DrawerTarget {
  number: string;
  page: string | null;
  city: string | null;
}

function withFilters(params: URLSearchParams, target: DrawerTarget): URLSearchParams {
  if (target.page) params.set("page", target.page);
  if (target.city) params.set("city", target.city);
  return params;
}

/** URL used to fetch the drawer partial, preserving the active page/city filters. */
export function buildPartialUrl(target: DrawerTarget): string {
  const qs = withFilters(new URLSearchParams({ partial: "1" }), target);
  return `/id/${target.number}?${qs.toString()}`;
}

/** History URL pushed when a drawer opens (params encoded, never raw spaces). */
export function buildHistoryUrl(target: DrawerTarget): string {
  const search = withFilters(new URLSearchParams(), target).toString();
  return search ? `/id/${target.number}?${search}` : `/id/${target.number}`;
}

/** Extract a target from an `/id/:number` href, or null if it isn't one. */
export function parseDrawerHref(
  href: string,
  origin: string = window.location.origin
): DrawerTarget | null {
  const match = href.match(/\/id\/(\d+)/);
  if (!match) return null;
  const url = new URL(href, origin);
  return {
    number: match[1],
    page: url.searchParams.get("page"),
    city: url.searchParams.get("city"),
  };
}

/** Read the current target from a location, or null when not on an id route. */
export function parseDrawerLocation(loc: {
  pathname: string;
  search: string;
}): DrawerTarget | null {
  const match = loc.pathname.match(/\/id\/(\d+)/);
  if (!match) return null;
  const params = new URLSearchParams(loc.search);
  return {
    number: match[1],
    page: params.get("page"),
    city: params.get("city"),
  };
}
