import { KeyRound, Mail, Smartphone } from "lucide-react";
import { DocCallout } from "@/features/web/docs/primitives/doc-callout";
import { DocFeatureGrid } from "@/features/web/docs/primitives/doc-feature-grid";
import { DocKeyValue } from "@/features/web/docs/primitives/doc-key-value";
import { DocSection } from "@/features/web/docs/primitives/doc-section";
import { DocSteps } from "@/features/web/docs/primitives/doc-steps";

export default function SignupSignin() {
  return (
    <>
      <DocSection id="three-ways" title="三种注册方式">
        <p>
          斗学提供三种注册和登录方式，选一种最顺手的即可——它们都能绑定到同一个账号上。
        </p>
        <DocFeatureGrid
          columns={3}
          items={[
            {
              icon: Mail,
              iconColor: "text-teal-600",
              title: "邮箱验证码",
              desc: "输入邮箱后获取 6 位数字验证码，填入即可完成注册或登录。",
            },
            {
              icon: KeyRound,
              iconColor: "text-blue-600",
              title: "账号密码",
              desc: "用设置好的用户名或邮箱，配合密码登录。适合已经注册过的老用户。",
            },
            {
              icon: Smartphone,
              iconColor: "text-amber-600",
              title: "微信扫码",
              desc: "用微信&ldquo;扫一扫&rdquo;快速注册或登录，不需要记账号密码。",
            },
          ]}
        />
      </DocSection>

      <DocSection id="rules" title="账号规则">
        <p>填写信息时请注意下面几条规则，不满足会被系统拦下来：</p>
        <DocKeyValue
          items={[
            {
              key: "用户名",
              value: "最多 30 字符",
              note: "只允许字母、数字、下划线 _、连字符 -",
            },
            {
              key: "密码",
              value: "至少 8 个字符",
              note: "必须同时包含大写字母、小写字母和数字",
            },
            {
              key: "邮箱验证码",
              value: "6 位纯数字",
              note: "发送后有一个短暂的倒计时才能再次发送",
            },
          ]}
        />
      </DocSection>

      <DocSection id="forgot" title="忘记密码怎么办">
        <p>
          如果忘记密码，不需要联系客服——在登录页有&ldquo;忘记密码&rdquo;的入口，走邮箱验证码流程即可自助重置。
        </p>
        <DocSteps
          steps={[
            {
              title: "进入忘记密码页",
              desc: "在登录页点击&ldquo;忘记密码&rdquo;链接",
            },
            {
              title: "获取邮箱验证码",
              desc: "输入注册时使用的邮箱，点击发送，等待 6 位数字验证码",
            },
            {
              title: "设置新密码",
              desc: "输入新密码（遵循上面的密码规则），确认后即可用新密码登录",
            },
          ]}
        />
      </DocSection>

      <DocSection id="invited" title="被朋友邀请注册">
        <p>
          如果你是通过朋友分享的邀请链接来到斗学的，完成注册后这个邀请关系会自动建立，朋友会在自己的推广页看到你出现——你不需要额外填写任何&ldquo;推荐人&rdquo;字段。
        </p>
        <DocCallout variant="tip" title="全程自动">
          邀请关系通过浏览器 cookie 识别。只要你是通过邀请链接打开的注册页，系统就会自动记录，不会因为你中途关闭页面而丢失。
        </DocCallout>
      </DocSection>
    </>
  );
}
