import { DocSection } from "@/features/web/docs/primitives/doc-section";
import { DocCallout } from "@/features/web/docs/primitives/doc-callout";
import { LegalPlaceholder } from "@/features/com/legal/components/legal-placeholder";
import { LegalPlaceholderNotice } from "@/features/com/legal/components/legal-placeholder-notice";
import { AgreementLink } from "@/features/com/legal/components/agreement-link";
import {
  BRAND,
  DOMAIN,
  EFFECTIVE_DATE,
  LAST_UPDATED,
  PLACEHOLDERS,
} from "@/features/com/legal/constants";

export function GuardianConsentDoc() {
  return (
    <>
      <LegalPlaceholderNotice fields={["companyName", "supportEmail"]} />
      <p className="text-xs text-slate-500">
        生效日期：{EFFECTIVE_DATE} · 最近更新：{LAST_UPDATED}
      </p>

      {/* Addressing paragraphs */}
      <p>尊敬的家长 / 监护人：</p>
      <p>
        感谢您对 {BRAND}（域名：{DOMAIN}
        ）的关注与支持。为保护未成年人的合法权益，根据《中华人民共和国未成年人保护法》《中华人民共和国个人信息保护法》《中华人民共和国未成年人网络保护条例》等相关法律法规的要求，我们特制定本《监护人同意书》。在您的孩子注册或使用本平台服务前，请您仔细阅读并理解以下内容。本平台运营主体为{" "}
        <LegalPlaceholder>{PLACEHOLDERS.companyName}</LegalPlaceholder>。
      </p>

      {/* clause-1 */}
      <DocSection id="clause-1" title="第1条 服务说明">
        <p>
          1.1 {BRAND}
          是一款面向英语学习者的在线学习平台，提供包括但不限于：
        </p>
        <ul className="list-disc space-y-2 pl-5">
          <li>英语学习课程与游戏化练习（单人、PK、小组等多种模式）</li>
          <li>学习进度跟踪与数据可视化</li>
          <li>在线学习社区（斗学社）与学习小组互动功能</li>
          <li>AI 辅助的个性化学习服务（如 AI 随心学）</li>
        </ul>
        <p>1.2 本平台致力于为用户提供安全、健康、有益的学习环境。</p>
      </DocSection>

      {/* clause-2 */}
      <DocSection id="clause-2" title="第2条 未成年人信息收集与使用">
        <p>2.1 在您的孩子使用本平台服务时，我们可能会收集以下信息：</p>
        <ul className="list-disc space-y-2 pl-5">
          <li>
            2.1.1 注册信息：用户名、电子邮箱、微信 openid/unionid（微信快捷登录场景）。
          </li>
          <li>
            2.1.2 学习数据：学习进度、练习记录、生词 / 复习 /
            已掌握记录、成就与排行榜数据。
          </li>
          <li>
            2.1.3 设备信息：设备型号、操作系统、浏览器类型、IP 地址。
          </li>
          <li>
            2.1.4 用户生成内容：学习笔记、上传的学习素材、在社区 / 小组发布的内容。
          </li>
        </ul>

        <p>2.2 我们收集这些信息的目的是：</p>
        <ul className="list-disc space-y-2 pl-5">
          <li>提供个性化的学习服务</li>
          <li>改进产品功能和用户体验</li>
          <li>保障账户安全</li>
          <li>履行法律法规要求</li>
        </ul>

        <p>
          2.3 <strong>我们承诺</strong>：
        </p>
        <ul className="list-disc space-y-2 pl-5">
          <li>2.3.1 仅收集提供服务所必需的最少信息；</li>
          <li>
            2.3.2 <strong>不会将未成年人信息用于商业营销；</strong>
          </li>
          <li>2.3.3 采取严格的安全措施保护未成年人信息；</li>
          <li>
            2.3.4{" "}
            <strong>未经监护人同意，不会向第三方提供未成年人信息</strong>。
          </li>
        </ul>
      </DocSection>

      {/* clause-3 */}
      <DocSection id="clause-3" title="第3条 监护人的权利">
        <p>3.1 作为监护人，您有权：</p>
        <ul className="list-disc space-y-2 pl-5">
          <li>3.1.1 随时访问、更正或删除您孩子的个人信息；</li>
          <li>3.1.2 要求我们停止收集或使用您孩子的个人信息；</li>
          <li>3.1.3 撤回您的同意，并要求注销账户；</li>
          <li>3.1.4 了解我们如何收集、使用和保护您孩子的信息。</li>
        </ul>

        <p>
          3.2 如需行使上述权利，请通过{" "}
          <LegalPlaceholder>{PLACEHOLDERS.supportEmail}</LegalPlaceholder> 联系我们。更多关于个人信息权利的说明，详见{" "}
          <AgreementLink slug="privacy-policy" />。
        </p>
      </DocSection>

      {/* clause-4 */}
      <DocSection id="clause-4" title="第4条 监护人的责任">
        <p>4.1 作为监护人，您应当：</p>
        <ul className="list-disc space-y-2 pl-5">
          <li>4.1.1 指导您的孩子正确、安全地使用本平台服务；</li>
          <li>
            4.1.2 监督您的孩子在本平台上的行为，确保遵守相关规则；
          </li>
          <li>4.1.3 妥善保管账户信息，防止账户被他人非法使用；</li>
          <li>4.1.4 定期关注您孩子的学习情况和网络使用时间。</li>
        </ul>

        <p>
          4.2 若您发现您的孩子在使用本平台时存在不当行为或遇到问题，请及时通过{" "}
          <LegalPlaceholder>{PLACEHOLDERS.supportEmail}</LegalPlaceholder> 与我们联系。
        </p>
      </DocSection>

      {/* clause-5 */}
      <DocSection id="clause-5" title="第5条 内容安全保障">
        <p>
          5.1 我们承诺为未成年人提供健康、积极的学习内容，并采取以下措施：
        </p>
        <ul className="list-disc space-y-2 pl-5">
          <li>
            5.1.1 对平台内容（含社区、小组 UGC）进行严格审核，过滤不良信息；
          </li>
          <li>5.1.2 建立举报机制，及时处理违规内容；</li>
          <li>5.1.3 限制未成年人访问不适宜的功能或内容；</li>
          <li>5.1.4 定期开展网络安全教育。</li>
        </ul>

        <DocCallout variant="info" title="内容举报渠道">
          <p>
            如发现平台内存在不适合未成年人的内容，请通过页面上的举报功能或发送邮件至{" "}
            <LegalPlaceholder>{PLACEHOLDERS.supportEmail}</LegalPlaceholder>{" "}
            向我们反馈，我们将在收到举报后 24 小时内予以处理。
          </p>
        </DocCallout>
      </DocSection>

      {/* clause-6 */}
      <DocSection id="clause-6" title="第6条 使用时长管理">
        <p>6.1 为保护未成年人身心健康，我们建议：</p>
        <ul className="list-disc space-y-2 pl-5">
          <li>6.1.1 合理安排学习时间，避免长时间连续使用；</li>
          <li>
            6.1.2 每学习 45-60 分钟，建议休息 10-15 分钟；
          </li>
          <li>
            6.1.3 监护人可通过关注账户的学习时长数据（在 &ldquo;我的主页&rdquo;
            可见），协助管理孩子的使用时长。
          </li>
        </ul>

        <p>
          6.2 本平台依据《中华人民共和国未成年人网络保护条例》对未成年用户实施使用时长管理，相关限制规则以平台公告为准。如监护人认为孩子的使用时长超出合理范围，可通过{" "}
          <LegalPlaceholder>{PLACEHOLDERS.supportEmail}</LegalPlaceholder> 联系我们申请协助管理。
        </p>
      </DocSection>

      {/* clause-7 */}
      <DocSection id="clause-7" title="第7条 付费服务说明">
        <p>7.1 本平台提供部分付费服务，包括但不限于：</p>
        <ul className="list-disc space-y-2 pl-5">
          <li>7.1.1 会员服务（月度 / 季度 / 年度 / 终身等档位）；</li>
          <li>7.1.2 能量豆充值（用于 AI 随心学等功能）。</li>
        </ul>

        <p>
          7.2{" "}
          <strong>
            未成年人购买付费服务，需经监护人同意并由监护人完成支付。
          </strong>
        </p>

        <p>
          7.3 我们建议监护人定期检查账户消费记录，如有疑问请及时通过{" "}
          <LegalPlaceholder>{PLACEHOLDERS.supportEmail}</LegalPlaceholder> 与我们联系。
        </p>

        <p>
          关于付费服务的完整条款，另见 <AgreementLink slug="product-service" />。
        </p>

        <DocCallout variant="warning" title="未成年人消费提示">
          <p>
            本平台不支持未成年人独立完成付费操作。若发现未成年人未经监护人同意擅自发起付费，请及时通过{" "}
            <LegalPlaceholder>{PLACEHOLDERS.supportEmail}</LegalPlaceholder>{" "}
            联系我们，我们将按照相关法律法规的规定协助处理。
          </p>
        </DocCallout>
      </DocSection>

      {/* clause-8 */}
      <DocSection id="clause-8" title="第8条 信息安全保护">
        <p>
          8.1 我们采取以下技术和管理措施保护未成年人信息安全：
        </p>
        <ul className="list-disc space-y-2 pl-5">
          <li>8.1.1 数据加密传输和存储；</li>
          <li>8.1.2 访问权限控制；</li>
          <li>8.1.3 定期安全审计；</li>
          <li>8.1.4 应急响应机制。</li>
        </ul>

        <p>
          8.2 若发生信息泄露等安全事件，我们将立即采取补救措施，并及时通知监护人。
        </p>

        <p>
          8.3 本平台对未成年人的个人信息采取以下额外保护措施：
        </p>
        <ul className="list-disc space-y-2 pl-5">
          <li>
            对内部员工访问未成年人信息实施更严格的权限管理，仅限履行职责所必需的人员方可访问；
          </li>
          <li>
            不对未成年用户进行个人信息画像或定向商业推广；
          </li>
          <li>
            未成年人的学习记录和账户信息不用于任何形式的商业目的，包括但不限于数据分析变现、广告投放、第三方数据共享等。
          </li>
        </ul>
      </DocSection>

      {/* clause-9 */}
      <DocSection id="clause-9" title="第9条 同意书的变更">
        <p>
          9.1 我们可能会根据法律法规变化或业务发展需要更新本同意书。
        </p>
        <p>
          9.2 重大变更将通过平台公告、邮件或站内消息等方式通知监护人。
        </p>
        <p>
          9.3 若您不同意变更后的内容，可以选择停止使用服务并注销账户。
        </p>
        <p>
          9.4 继续允许您的孩子使用本平台服务，即视为您已接受更新后的同意书内容。
        </p>
      </DocSection>

      {/* clause-10 */}
      <DocSection id="clause-10" title="第10条 联系我们">
        <p>
          10.1 如果您对本同意书有任何疑问，或需要行使监护人权利，请通过{" "}
          <LegalPlaceholder>{PLACEHOLDERS.supportEmail}</LegalPlaceholder> 联系我们。
        </p>
        <p>我们将在收到您的请求后 15 个工作日内予以回复。</p>

        <p>10.2 您可通过上述联系方式处理以下事项：</p>
        <ul className="list-disc space-y-2 pl-5">
          <li>查询、更正或删除您孩子的个人信息；</li>
          <li>撤回对本同意书的同意并申请注销账户；</li>
          <li>反馈平台内容安全问题；</li>
          <li>申请协助管理未成年人使用时长；</li>
          <li>就付费服务相关事项提出咨询或异议。</li>
        </ul>

        <p>
          10.3 若您对本平台处理未成年人个人信息的行为存在异议，可向国家互联网信息办公室或您所在地的有关监管部门投诉举报。
        </p>
      </DocSection>

      {/* clause-11 */}
      <DocSection id="clause-11" title="第11条 监护人声明">
        <p>11 通过勾选同意本《监护人同意书》，您确认：</p>
        <ul className="list-disc space-y-2 pl-5">
          <li>11.1 您是该未成年用户的合法监护人；</li>
          <li>
            11.2 您已仔细阅读并完全理解本同意书的全部内容；
          </li>
          <li>
            11.3 您同意您的孩子注册并使用 {BRAND}（域名：{DOMAIN}）的服务；
          </li>
          <li>
            11.4 您同意我们按照本同意书的约定收集、使用和保护您孩子的个人信息；
          </li>
          <li>11.5 您将履行监护人的监督和指导责任。</li>
        </ul>

        <p>
          本同意书的最终解释权归 <LegalPlaceholder>{PLACEHOLDERS.companyName}</LegalPlaceholder>。
        </p>
      </DocSection>
    </>
  );
}
