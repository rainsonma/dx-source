import { DocCallout } from "@/features/web/docs/primitives/doc-callout";
import { DocSection } from "@/features/web/docs/primitives/doc-section";
import { DocSteps } from "@/features/web/docs/primitives/doc-steps";

export default function Security() {
  return (
    <>
      <DocSection id="change-email" title="修改邮箱">
        <DocSteps
          steps={[
            {
              title: "进入账号安全页",
              desc: "在个人中心找到&ldquo;账号安全&rdquo;设置",
            },
            {
              title: "输入新邮箱",
              desc: "系统会向新邮箱发送 6 位数字验证码",
            },
            {
              title: "输入验证码确认",
              desc: "验证通过后新邮箱立即生效，旧邮箱失效",
            },
          ]}
        />
      </DocSection>

      <DocSection id="change-password" title="修改密码">
        <p>
          在账号安全页面点&ldquo;修改密码&rdquo;。你需要先输入当前密码（作为身份验证），然后输入新密码。新密码必须符合斗学的密码规则：至少 8 个字符，必须同时包含大写字母、小写字母和数字。
        </p>
      </DocSection>

      <DocSection id="logout" title="登出">
        <p>
          登出会让当前设备的登录状态失效——下次访问斗学需要重新登录。登出本身是完全安全的动作，你的数据不受影响，只是当前浏览器的登录凭证被清除了。
        </p>
      </DocSection>

      <DocSection id="session-replaced" title="多设备登录：新设备会踢掉旧设备">
        <DocCallout variant="warning" title="同一账号同时只能在一处登录">
          斗学对每个账号只允许一处活跃会话——当你在新设备上登录时，旧设备的登录凭证会立即失效。旧设备再次操作会看到&ldquo;会话已在别处登录&rdquo;的提示，你可以在旧设备上重新登录，但同样会踢掉更前面的会话。
        </DocCallout>
      </DocSection>
    </>
  );
}
