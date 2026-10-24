import type { ReactNode } from "react";
import { hrefOf, navigate, type Route } from "../lib/router";

export function RouteLink({
  route,
  className,
  children,
}: {
  route: Route;
  className?: string;
  children: ReactNode;
}) {
  return (
    <a
      className={className}
      href={hrefOf(route)}
      onClick={(e) => {
        if (e.metaKey || e.ctrlKey || e.shiftKey || e.button !== 0) return;
        e.preventDefault();
        navigate(route);
      }}
    >
      {children}
    </a>
  );
}
