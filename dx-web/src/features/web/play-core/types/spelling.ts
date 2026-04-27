export type SpellingItem = {
  item: string;
  answer: boolean;
  pos: string | null;
  position: number;
  definition: string;
  phonetic: { uk: string; us: string } | null;
};

export type TypedWord = {
  text: string;
  isAnswer: boolean;
};
