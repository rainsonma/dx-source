"use client";

import { useState, useCallback } from "react";
import { HeroCard } from "@/features/web/games/components/hero-card";
import { LevelGrid } from "@/features/web/games/components/level-grid";
import { GameModeCard } from "@/features/web/play-core/components/game-mode-card";
import { useFavoriteToggle } from "@/features/web/games/hooks/use-favorite-toggle";

type Level = { id: string; name: string; order: number };

interface GameDetailContentProps {
  game: {
    id: string;
    name: string;
    description: string;
    mode: string;
    coverUrl: string | null;
    levelCount: number;
    playerCount: string;
    levels: Level[];
    completedLevels: number;
    isVip: boolean;
  };
  heroSession: {
    degree: string;
    pattern: string | null;
    levelName: string;
  } | null;
  isFavorited: boolean;
  rules: React.ReactNode;
  stats: React.ReactNode;
}

export function GameDetailContent({
  game,
  heroSession,
  isFavorited,
  rules,
  stats,
}: GameDetailContentProps) {
  const [modalOpen, setModalOpen] = useState(false);
  const [modalKey, setModalKey] = useState(0);
  const [modalMode, setModalMode] = useState<"single" | "pk">("single");
  const [selectedLevel, setSelectedLevel] = useState<{
    id: string;
    label: string;
  } | null>(null);

  const handleStart = useCallback(() => {
    setSelectedLevel(null);
    setModalMode("single");
    setModalKey((k) => k + 1);
    setModalOpen(true);
  }, []);

  const handlePkStart = useCallback(() => {
    setSelectedLevel(null);
    setModalMode("pk");
    setModalKey((k) => k + 1);
    setModalOpen(true);
  }, []);

  const handleLevelClick = useCallback((levelId: string, levelName: string) => {
    setSelectedLevel({ id: levelId, label: levelName });
    setModalKey((k) => k + 1);
    setModalOpen(true);
  }, []);

  const { favorited, toggle, isPending: isFavoritePending } =
    useFavoriteToggle(game.id, game.name, isFavorited);

  return (
    <>
      <HeroCard
        id={game.id}
        title={game.name}
        description={game.description}
        mode={game.mode}
        levelCount={game.levelCount}
        playerCount={game.playerCount}
        coverUrl={game.coverUrl}
        resumeLabel={heroSession?.levelName ?? null}
        onStart={handleStart}
        onPkStart={handlePkStart}
        isFavorited={favorited}
        onFavoriteToggle={toggle}
        isFavoritePending={isFavoritePending}
      />

      <div className="flex flex-1 flex-col gap-5 lg:flex-row">
        <div className="flex-1">
          <LevelGrid
            levels={game.levels}
            completedLevels={game.completedLevels}
            isVip={game.isVip}
            onLevelClick={handleLevelClick}
          />
        </div>
        <div className="flex w-full flex-col gap-5 lg:w-80 lg:shrink-0">
          {rules}
          {stats}
        </div>
      </div>

      <GameModeCard
        key={modalKey}
        gameId={game.id}
        gameName={game.name}
        gameMode={game.mode}
        mode={modalMode}
        levelId={selectedLevel?.id}
        levelLabel={selectedLevel?.label}
        initialDegree={heroSession?.degree}
        initialPattern={heroSession?.pattern}
        open={modalOpen}
        onClose={() => setModalOpen(false)}
      />
    </>
  );
}
