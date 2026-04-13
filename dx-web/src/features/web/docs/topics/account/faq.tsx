import { DocFaqAccordion } from "@/features/web/docs/primitives/doc-faq-accordion";
import { DocSection } from "@/features/web/docs/primitives/doc-section";

export default function Faq() {
  return (
    <>
      <DocSection id="account" title="账号与登录">
        <p>遇到账号或登录问题？先看看下面几条常见情况。</p>
        <DocFaqAccordion
          items={[
            {
              question: "忘记密码怎么办？",
              answer:
                "在登录页点击&ldquo;忘记密码&rdquo;，通过邮箱验证码重设密码。如果注册时用的邮箱你也不记得了，请在反馈页提交 Bug 报告，附上你可能用过的邮箱地址。",
            },
            {
              question: "验证码为什么收不到？",
              answer:
                "先检查垃圾邮件和垃圾箱；如果仍没收到，稍等几分钟后重发。某些邮箱服务商可能会延迟 5-10 分钟。如果尝试了几次都失败，换个邮箱地址再试。",
            },
            {
              question: "微信扫码登录失败？",
              answer:
                "请确保微信是最新版，网络正常，扫码时手机和电脑处于同一网络环境。如果反复失败，改用邮箱验证码登录——效果是一样的。",
            },
            {
              question: "为什么我被踢下线了？",
              answer:
                "斗学对每个账号同一时间只允许一处活跃会话。如果有人（或你自己在别的设备上）登录了你的账号，当前会话会被踢掉。重新登录即可，但请确认你的密码是否泄露。",
            },
          ]}
        />
      </DocSection>

      <DocSection id="payment" title="会员与支付">
        <p>关于会员购买、订单状态、支付问题的常见疑问。</p>
        <DocFaqAccordion
          items={[
            {
              question: "订单支付后会员没到账怎么办？",
              answer:
                "正常情况下支付完成 1 分钟内会员就会生效。如果超过 10 分钟仍未到账，请在反馈里提交&ldquo;Bug 报告&rdquo;并附上订单号，我们会帮你排查。",
            },
            {
              question: "订单多久会过期？",
              answer:
                "订单创建后 30 分钟内如果没有完成支付，会自动变为过期状态，不能再继续支付。你需要回到购买页重新生成一笔新订单。",
            },
            {
              question: "可以退款吗？",
              answer:
                "数字商品一经发放不支持退款，请在购买前确认清楚需要的档位。如果遇到系统错误导致的重复扣款，在反馈里提交 Bug 报告。",
            },
            {
              question: "终身会员真的是永久吗？",
              answer:
                "是的。终身会员一次购买永久有效，不需要续费，并且每月会继续赠送 15,000 能量豆——这一点比其它档位多。",
            },
          ]}
        />
      </DocSection>

      <DocSection id="learning" title="游戏与学习">
        <p>玩游戏、闯关、得经验相关的常见问题。</p>
        <DocFaqAccordion
          items={[
            {
              question: "为什么这关被锁住了？",
              answer:
                "免费用户只能玩每个游戏的第一关；其余关卡需要 VIP 会员才能解锁。开通会员后所有关卡立即开放。",
            },
            {
              question: "学习数据为什么不同步？",
              answer:
                "先刷新页面一次。如果仍不同步，检查你的网络连接，或者确认你没有被多设备登录机制踢下线。如果问题持续，请在反馈里告诉我们具体情况。",
            },
            {
              question: "PK 模式连不上对方怎么办？",
              answer:
                "指定对手的 PK 要求对方是 VIP 会员且当前在线。如果对方刚好离线或者网络不稳，邀请会失败。可以改用随机匹配，系统会自动为你配对，立刻就能开始。",
            },
            {
              question: "为什么要 60% 正确率才给经验？",
              answer:
                "这是为了保证经验反映真实的掌握度。如果低于 60% 就发经验，你可能会习惯性乱点来刷数据，对学习反而是反效果。60% 是一个&ldquo;允许犯错但要求用心&rdquo;的阈值。",
            },
          ]}
        />
      </DocSection>

      <DocSection id="ai" title="AI 与能量豆">
        <p>AI 随心学和能量豆相关的常见疑问。</p>
        <DocFaqAccordion
          items={[
            {
              question: "能量豆不够用怎么办？",
              answer:
                "月/季/年会员每月自动赠送 10,000 豆，终身会员每月 15,000 豆。如果用完了还想继续，可以直接在充值页购买——五档充值包任选，10 元包开始有赠送比例。",
            },
            {
              question: "AI 生成失败了，豆会退吗？",
              answer:
                "会的。AI 生成如果因为网络、模型异常等原因失败，消耗的豆会自动退还到你的账户，不会白扣。如果你发现扣了豆但没退，请在反馈里告诉我们。",
            },
            {
              question: "能量豆会过期吗？",
              answer:
                "会员每月赠送的豆在下一次赠送时会被清零（即&ldquo;这个月的豆这个月用&rdquo;）。自己花钱购买的豆永不清零，可以随时使用。",
            },
          ]}
        />
      </DocSection>

      <DocSection id="tech" title="技术与兼容性">
        <p>浏览器、设备、网络相关的常见问题。</p>
        <DocFaqAccordion
          items={[
            {
              question: "斗学支持哪些浏览器？",
              answer:
                "Chrome、Safari、Edge 最新版都能获得最佳体验。Firefox 也可以用，但某些交互动画可能略有差异。旧版 IE 或很老的移动浏览器可能出现兼容问题。",
            },
            {
              question: "图片上传失败怎么办？",
              answer:
                "请确认文件小于 2MB，格式为 JPEG 或 PNG。超出限制的文件会被系统拒绝。如果你的图片过大，可以先用压缩工具或手机自带的&ldquo;调整大小&rdquo;功能处理一下。",
            },
            {
              question: "页面卡住不动是怎么回事？",
              answer:
                "先尝试刷新页面，清一下浏览器缓存再重新登录。如果问题持续，请在反馈里附上你使用的浏览器版本和具体的操作步骤，我们会尽快定位。",
            },
          ]}
        />
      </DocSection>
    </>
  );
}
