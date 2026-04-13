"use client";

import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion";

type Item = {
  question: string;
  answer: string;
};

type Props = {
  items: Item[];
};

export function DocFaqAccordion({ items }: Props) {
  return (
    <Accordion type="single" collapsible className="w-full">
      {items.map((item, i) => (
        <AccordionItem key={i} value={`item-${i}`}>
          <AccordionTrigger className="text-left text-[15px] font-medium text-slate-800">
            {item.question}
          </AccordionTrigger>
          <AccordionContent className="text-sm leading-[1.7] text-slate-600">
            {item.answer}
          </AccordionContent>
        </AccordionItem>
      ))}
    </Accordion>
  );
}
