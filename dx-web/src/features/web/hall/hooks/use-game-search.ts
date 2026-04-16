"use client";

import { useState, useEffect, useRef } from "react";
import type { GameSearchResult } from "@/features/web/hall/actions/game-search.action";
import {
  searchGamesAction,
  getRecentGamesAction,
} from "@/features/web/hall/actions/game-search.action";

/** Manages game search dialog state with debounced server queries */
export function useGameSearch() {
  const [isOpen, setIsOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<GameSearchResult[]>([]);
  const [recentGames, setRecentGames] = useState<GameSearchResult[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const recentLoadedRef = useRef(false);

  /* eslint-disable react-hooks/set-state-in-effect -- search dialog patterns require setState in effects */

  /** Load recent games once when dialog opens */
  useEffect(() => {
    if (!isOpen) return;
    if (recentLoadedRef.current) return;

    recentLoadedRef.current = true;
    getRecentGamesAction().then((res) => {
      if (!res.error) setRecentGames(res.games);
    });
  }, [isOpen]);

  /** Reset query when dialog closes */
  useEffect(() => {
    if (!isOpen) {
      setQuery("");
      setResults([]);
    }
  }, [isOpen]);

  /** Debounced search on query change */
  useEffect(() => {
    if (timerRef.current) clearTimeout(timerRef.current);

    const trimmed = query.trim();
    if (!trimmed) {
      setResults([]);
      setIsLoading(false);
      return;
    }

    let stale = false;
    setIsLoading(true);
    timerRef.current = setTimeout(async () => {
      const res = await searchGamesAction(trimmed);
      if (!stale) {
        setResults(res.games);
        setIsLoading(false);
      }
    }, 300);

    return () => {
      if (timerRef.current) clearTimeout(timerRef.current);
      stale = true;
    };
  }, [query]);

  /* eslint-enable react-hooks/set-state-in-effect */

  /** Register Cmd+K / Ctrl+K keyboard shortcut */
  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      if (e.key === "k" && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setIsOpen((prev) => !prev);
      }
    }

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, []);

  /** The items to display: search results when typing, recent games otherwise */
  const displayItems = query.trim() ? results : recentGames;
  const groupLabel = query.trim() ? "猜你想学" : "最近玩过";
  const showGroup = displayItems.length > 0;

  return {
    isOpen,
    setIsOpen,
    query,
    setQuery,
    displayItems,
    groupLabel,
    showGroup,
    isLoading,
  };
}
