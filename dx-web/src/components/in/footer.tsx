import { GraduationCap } from "lucide-react";

const footerColumns = [
  {
    title: "服务条款",
    links: ["用户协议", "隐私政策", "监护人同意书", "产品服务协议", "Cookie 政策"],
  },
  {
    title: "斗学产品",
    links: ["渐进学习法", "AI 智能定制", "多重游戏模式", "丰富课程体系", "社群小组"],
  },
  {
    title: "斗学团队",
    links: ["关于我们", "建议反馈", "内容投稿", "商务合作", "bs@douxue.cc"],
  },
];

export function Footer() {
  return (
    <footer id="contact" className="scroll-mt-20 w-full border-t border-slate-200 bg-slate-50">
      <div className="mx-auto flex max-w-[1280px] flex-col gap-12 px-[120px] pb-10 pt-[60px]">
        {/* Top section */}
        <div className="flex w-full flex-col gap-10 xl:flex-row xl:justify-between">
          {/* Brand */}
          <div className="flex flex-col gap-4">
            <div className="flex items-center gap-2.5">
              <GraduationCap className="h-7 w-7 text-teal-600" />
              <span className="text-lg font-extrabold text-slate-900">斗学</span>
            </div>
            <p className="max-w-[280px] text-sm leading-[1.5] text-slate-500">
              玩游戏，学英语，AI 智能辅助，斗学重新定义英语学习体验，让进步自然发生...
            </p>
          </div>

          {/* Link columns */}
          <div className="grid grid-cols-1 gap-10 md:grid-cols-2 lg:grid-cols-3 xl:flex xl:gap-16">
            {footerColumns.map((col) => (
              <div key={col.title} className="flex flex-col gap-4">
                <h4 className="text-[13px] font-semibold tracking-[1px] text-slate-900">
                  {col.title}
                </h4>
                {col.links.map((link) => (
                  <span
                    key={link}
                    className="cursor-pointer text-sm text-slate-500 hover:text-slate-700"
                  >
                    {link}
                  </span>
                ))}
              </div>
            ))}
          </div>

          {/* Contact column */}
          <div className="flex flex-col items-start gap-4 xl:items-end">
            <h4 className="text-[13px] font-semibold tracking-[1px] text-slate-900">
              联系我们
            </h4>
            <div className="flex h-[140px] w-[140px] items-center justify-center rounded-lg bg-slate-100">
              <span className="text-xs text-slate-400">微信二维码</span>
            </div>
            <span className="text-xs text-slate-400">微信扫一扫联系小助手</span>
          </div>
        </div>

        {/* Divider */}
        <div className="h-px w-full bg-slate-200" />

        {/* Bottom copyright */}
        <div className="flex w-full flex-col items-center gap-2">
          <span className="text-[13px] text-slate-400">
            © 2026 douxue.cc 版权所有
          </span>
          <span className="text-[13px] text-slate-400">
            京公网安备 xxxxxxxxxxxxxx 号  京 ICP 备 xxxxxxxxxx 号
          </span>
        </div>
      </div>
    </footer>
  );
}
