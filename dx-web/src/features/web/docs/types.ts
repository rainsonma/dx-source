import type { ComponentType } from "react";
import type { LucideIcon } from "lucide-react";

export type DocTopic = {
  slug: string;
  title: string;
  description: string;
  Component: ComponentType;
  groupLabel?: string;
};

export type DocCategory = {
  slug: string;
  title: string;
  description: string;
  icon: LucideIcon;
  accentClass: string;
  topics: DocTopic[];
};

export type TopicRef = {
  category: DocCategory;
  topic: DocTopic;
};
