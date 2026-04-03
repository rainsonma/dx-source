export type BeanPackage = {
  slug: string;
  beans: number;
  bonus: number;
  price: number;
  tag?: string;
};

export const BEAN_PACKAGES: BeanPackage[] = [
  { slug: "beans-1", beans: 1000, bonus: 0, price: 100 },
  { slug: "beans-5", beans: 5000, bonus: 0, price: 500 },
  { slug: "beans-10", beans: 10000, bonus: 1000, price: 1000, tag: "超值推荐" },
  { slug: "beans-50", beans: 50000, bonus: 7500, price: 5000 },
  { slug: "beans-100", beans: 100000, bonus: 20000, price: 10000, tag: "最划算" },
];
