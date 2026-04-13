"use client";

import { GAME_MODE_LABELS } from "@/consts/game-mode";

type CategoryOption = { id: string; name: string; depth: number; isLeaf: boolean };
type PressOption = { id: string; name: string };

type Filters = {
  categoryIds?: string[];
  pressId?: string;
  mode?: string;
};

type FilterSectionProps = {
  categories: CategoryOption[];
  presses: PressOption[];
  filters: Filters;
  onFiltersChange: (filters: Filters) => void;
  showPresses?: boolean;
};

function ActivePill({ label, onClick }: { label: string; onClick?: () => void }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="rounded-md bg-teal-600 px-3 py-1 text-xs font-semibold text-white"
    >
      {label}
    </button>
  );
}

function InactivePill({ label, onClick }: { label: string; onClick?: () => void }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="cursor-pointer text-xs text-muted-foreground hover:text-foreground"
    >
      {label}
    </button>
  );
}

export function FilterSection({
  categories,
  presses,
  filters,
  onFiltersChange,
  showPresses = true,
}: FilterSectionProps) {
  const topCategories = categories.filter((c) => c.depth === 0);
  const modeEntries = Object.entries(GAME_MODE_LABELS);

  function getChildrenOf(parentId: string): CategoryOption[] {
    const parentIdx = categories.findIndex((c) => c.id === parentId);
    if (parentIdx === -1) return [];
    const children: CategoryOption[] = [];
    for (let i = parentIdx + 1; i < categories.length; i++) {
      if (categories[i].depth === 0) break;
      children.push(categories[i]);
    }
    return children;
  }

  function findActiveTopId(): string | undefined {
    if (!filters.categoryIds || filters.categoryIds.length === 0) return undefined;
    const firstId = filters.categoryIds[0];
    if (topCategories.some((c) => c.id === firstId)) return firstId;
    for (const top of topCategories) {
      const children = getChildrenOf(top.id);
      if (children.some((ch) => ch.id === firstId)) return top.id;
    }
    return undefined;
  }

  const activeTopId = findActiveTopId();
  const childCategories = activeTopId ? getChildrenOf(activeTopId) : [];
  const hasChildren = childCategories.length > 0;

  const isAllWithinParent =
    activeTopId && filters.categoryIds && filters.categoryIds.length > 1;

  const selectedChildId =
    activeTopId &&
    filters.categoryIds?.length === 1 &&
    !topCategories.some((c) => c.id === filters.categoryIds![0])
      ? filters.categoryIds[0]
      : undefined;

  function handleTopCategoryClick(catId: string) {
    const children = getChildrenOf(catId);
    if (children.length === 0) {
      onFiltersChange({ ...filters, categoryIds: [catId] });
    } else {
      const allIds = [catId, ...children.map((c) => c.id)];
      onFiltersChange({ ...filters, categoryIds: allIds });
    }
  }

  function handleChildCategoryClick(childId: string) {
    onFiltersChange({ ...filters, categoryIds: [childId] });
  }

  function handleAllChildrenClick() {
    if (!activeTopId) return;
    const children = getChildrenOf(activeTopId);
    const allIds = [activeTopId, ...children.map((c) => c.id)];
    onFiltersChange({ ...filters, categoryIds: allIds });
  }

  return (
    <div className="flex w-full flex-col gap-3 rounded-xl border border-border bg-card px-4 py-3 lg:px-5 lg:py-4">
      {/* Category row */}
      <div className="flex w-full flex-wrap items-center gap-3 lg:gap-4">
        {!filters.categoryIds || filters.categoryIds.length === 0 ? (
          <ActivePill label="全部分类" />
        ) : (
          <InactivePill
            label="全部分类"
            onClick={() => onFiltersChange({ ...filters, categoryIds: undefined })}
          />
        )}
        {topCategories.map((cat) =>
          activeTopId === cat.id ? (
            <ActivePill key={cat.id} label={cat.name} />
          ) : (
            <InactivePill
              key={cat.id}
              label={cat.name}
              onClick={() => handleTopCategoryClick(cat.id)}
            />
          )
        )}
      </div>

      {/* Child category sub-row */}
      {hasChildren && (
        <div className="flex w-full flex-wrap items-center gap-3 pl-4 lg:gap-4">
          {isAllWithinParent ? (
            <ActivePill label="全部" />
          ) : (
            <InactivePill label="全部" onClick={handleAllChildrenClick} />
          )}
          {childCategories.map((cat) =>
            selectedChildId === cat.id ? (
              <ActivePill key={cat.id} label={cat.name} />
            ) : (
              <InactivePill
                key={cat.id}
                label={cat.name}
                onClick={() => handleChildCategoryClick(cat.id)}
              />
            )
          )}
        </div>
      )}

      <div className="h-px w-full bg-border" />

      {showPresses && (
        <>
          {/* Publisher row */}
          <div className="flex w-full flex-wrap items-center gap-4">
            {!filters.pressId ? (
              <ActivePill label="全部出版社" />
            ) : (
              <InactivePill
                label="全部出版社"
                onClick={() => onFiltersChange({ ...filters, pressId: undefined })}
              />
            )}
            {presses.map((press) =>
              filters.pressId === press.id ? (
                <ActivePill key={press.id} label={press.name} />
              ) : (
                <InactivePill
                  key={press.id}
                  label={press.name}
                  onClick={() => onFiltersChange({ ...filters, pressId: press.id })}
                />
              )
            )}
          </div>

          <div className="h-px w-full bg-border" />
        </>
      )}

      {/* Game mode row */}
      <div className="flex w-full flex-wrap items-center gap-3 lg:gap-4">
        {!filters.mode ? (
          <ActivePill label="全部游戏模式" />
        ) : (
          <InactivePill
            label="全部游戏模式"
            onClick={() => onFiltersChange({ ...filters, mode: undefined })}
          />
        )}
        {modeEntries.map(([value, label]) =>
          filters.mode === value ? (
            <ActivePill key={value} label={label} />
          ) : (
            <InactivePill
              key={value}
              label={label}
              onClick={() => onFiltersChange({ ...filters, mode: value })}
            />
          )
        )}
      </div>
    </div>
  );
}
