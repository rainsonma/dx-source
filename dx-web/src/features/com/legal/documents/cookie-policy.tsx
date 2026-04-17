import { DocSection } from "@/features/web/docs/primitives/doc-section";
import { LegalPlaceholderNotice } from "@/features/com/legal/components/legal-placeholder-notice";
import { AgreementLink } from "@/features/com/legal/components/agreement-link";
import {
  BRAND,
  DOMAIN,
  EFFECTIVE_DATE,
  LAST_UPDATED,
  PLACEHOLDERS,
} from "@/features/com/legal/constants";

function P({ children }: { children: React.ReactNode }) {
  return (
    <span className="rounded bg-amber-50 px-1 py-0.5 font-mono text-[13px] text-amber-700">
      {children}
    </span>
  );
}

export function CookiePolicyDoc() {
  return (
    <>
      <LegalPlaceholderNotice fields={["companyName", "supportEmail"]} />
      <p className="text-xs text-slate-500">
        生效日期：{EFFECTIVE_DATE} · 最近更新：{LAST_UPDATED}
      </p>

      {/* Preamble */}
      <p>
        本 Cookie 政策说明{BRAND}（域名：{DOMAIN}，运营主体{" "}
        <P>{PLACEHOLDERS.companyName}</P>
        ）在其网站及相关服务中，如何使用 Cookie、LocalStorage
        及其他类似的浏览器存储技术（以下统称
        &ldquo;Cookie 及类似技术&rdquo;）。我们建议您在使用本平台前通读本政策，更多关于个人信息处理的说明详见{" "}
        <AgreementLink slug="privacy-policy" />。
      </p>

      {/* clause-1 */}
      <DocSection id="clause-1" title="第1条 什么是 Cookie 及类似技术">
        <p>
          Cookie 是网站在您访问时，由服务器通过浏览器写入您本地设备的小型文本文件。Cookie
          中通常包含一个唯一标识符，用于在您多次访问同一网站时识别您的浏览器、保持登录状态、记住您的偏好设置，以及辅助统计分析。Cookie
          本身不包含您的姓名、密码等直接身份信息，但在结合其他数据时，可能与您的账号相关联。
        </p>
        <p>
          LocalStorage 与 SessionStorage
          是现代浏览器提供的本地存储机制，功能上与 Cookie
          类似，但容量更大，且数据完全保存在客户端，不会随每次网络请求自动发送至服务器。其中
          SessionStorage 在浏览器标签页关闭后自动清除，而 LocalStorage
          则会持久保存，直到您主动清除为止。
        </p>
        <p>
          本平台根据功能需要综合使用上述多种技术，统称为
          &ldquo;Cookie 及类似技术&rdquo;。这些技术是现代 Web
          应用正常运行所必需的基础设施，也是本平台为您提供安全、流畅学习体验的技术基础。
        </p>
      </DocSection>

      {/* clause-2 */}
      <DocSection id="clause-2" title="第2条 我们如何使用 Cookie">
        <p>
          <strong>2.1 使用目的分类</strong>
        </p>
        <ol className="list-decimal space-y-3 pl-5">
          <li>
            <strong>2.1.1 必要性 Cookie：</strong>
            用于维持登录态（如 <code>dx_token</code>{" "}
            凭证）、跨页面会话保持、表单防篡改与 CSRF
            防护。这类技术是本平台提供核心服务的前提，无这类 Cookie
            您将无法登录、购买会员或访问个人中心。
          </li>
          <li>
            <strong>2.1.2 功能性 Cookie：</strong>
            用于记忆您的偏好，如侧边栏折叠状态、主题选择、语言偏好（若适用）等界面设置。关闭这类
            Cookie 不会影响核心学习功能，但每次访问时您的界面偏好可能需要重新设置。
          </li>
          <li>
            <strong>2.1.3 分析性 Cookie：</strong>
            用于记录页面访问路径、学习行为汇总、功能使用频次，辅助本平台优化产品体验与排查服务故障。分析数据以汇总或匿名化形式使用，不用于
            &ldquo;大数据杀熟&rdquo; 或面向您个人的广告推送。
          </li>
        </ol>
        <p className="mt-3">
          <strong>2.2 不用于广告</strong>
        </p>
        <p>
          本平台当前不接入任何广告联盟或用户画像类追踪技术，不会将您的浏览行为用于跨站广告投放或商业画像分析。
        </p>
      </DocSection>

      {/* clause-3 */}
      <DocSection id="clause-3" title="第3条 Cookie 的分类">
        <p>下表列出本平台使用的主要 Cookie 及类似技术类别：</p>
        <table className="w-full border-collapse text-[13px]">
          <thead className="bg-slate-50">
            <tr>
              <th className="border border-slate-200 px-3 py-2 text-left font-semibold">
                类别
              </th>
              <th className="border border-slate-200 px-3 py-2 text-left font-semibold">
                示例
              </th>
              <th className="border border-slate-200 px-3 py-2 text-left font-semibold">
                目的
              </th>
              <th className="border border-slate-200 px-3 py-2 text-left font-semibold">
                存留时长
              </th>
              <th className="border border-slate-200 px-3 py-2 text-left font-semibold">
                是否可关闭
              </th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td className="border border-slate-200 px-3 py-2 align-top font-medium">
                必要性 Cookie
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                <code>dx_token</code> 登录凭证
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                维持登录状态、权限鉴权
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                会话 / 最长 7 天
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                不可关闭（关闭将无法登录）
              </td>
            </tr>
            <tr>
              <td className="border border-slate-200 px-3 py-2 align-top font-medium">
                必要性 Cookie
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                CSRF / 表单防护令牌
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                防止跨站请求伪造
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                会话
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                不可关闭
              </td>
            </tr>
            <tr>
              <td className="border border-slate-200 px-3 py-2 align-top font-medium">
                功能性 Cookie
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                UI 偏好（侧边栏、主题）
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                记忆您的界面选择
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                最长 30 天
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                可关闭（可能影响体验）
              </td>
            </tr>
            <tr>
              <td className="border border-slate-200 px-3 py-2 align-top font-medium">
                分析性 Cookie
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                页面访问路径、功能使用频次
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                产品优化、故障排查
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                最长 30 天
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                可关闭（不影响学习）
              </td>
            </tr>
          </tbody>
        </table>
      </DocSection>

      {/* clause-4 */}
      <DocSection id="clause-4" title="第4条 第三方 Cookie 与 SDK">
        <p>
          <strong>4.1</strong>{" "}
          当您使用第三方登录（如微信开放平台 OAuth）时，第三方服务方会在您的浏览器中设置其
          Cookie，用于完成其自身的身份验证流程。我们不主动读取这些跨域 Cookie，相关
          Cookie 的处理受该第三方服务方的 Cookie 政策约束。
        </p>
        <p>
          <strong>4.2</strong>{" "}
          当您通过平台支持的付费渠道完成支付时，支付服务方可能在支付流程中写入其
          Cookie。本平台仅与支付服务方共享必要的订单信息（详见{" "}
          <AgreementLink slug="privacy-policy" />{" "}
          第 4 条），不主动读取其 Cookie。
        </p>
        <p>
          <strong>4.3</strong>{" "}
          <strong>
            本平台当前不接入广告联盟 / 像素追踪 / 用户画像类 Cookie
          </strong>
          ，不会将您的浏览行为用于跨站广告投放。如本平台未来需要新增上述类别的第三方
          Cookie，将提前通过平台公告告知您，并在必要时取得您的同意。
        </p>
        <p>
          <strong>4.4</strong>{" "}
          对于第三方服务方写入的 Cookie，本平台不对其内容、安全性或准确性承担责任。您可通过各第三方服务方的官方隐私政策了解其
          Cookie 使用详情，并通过浏览器设置管理相关 Cookie。
        </p>
      </DocSection>

      {/* clause-5 */}
      <DocSection id="clause-5" title="第5条 您的选择与管理方式">
        <p>
          <strong>5.1</strong>{" "}
          您可以通过浏览器设置禁用 Cookie。
          <strong>
            禁用必要性 Cookie 将导致登录、购买、学习进度同步等核心功能不可用。
          </strong>
          在禁用前，请确保您了解这一影响。
        </p>
        <p>
          <strong>5.2</strong>{" "}
          主流浏览器（Chrome / Safari / Microsoft Edge / Firefox）均在其设置菜单中提供
          Cookie 管理入口；您可以进入浏览器的 &ldquo;隐私与安全&rdquo; 或
          &ldquo;Cookie 与站点数据&rdquo;
          设置，查看、删除或禁用本平台及第三方写入的 Cookie。具体操作路径请参阅对应浏览器的官方帮助文档。
        </p>
        <p>
          <strong>5.3</strong>{" "}
          如您已登录本平台，可在 &ldquo;个人主页&rdquo; → &ldquo;隐私设置&rdquo;{" "}
          查看并撤回对非必要信息收集的同意（详见{" "}
          <AgreementLink slug="privacy-policy" /> 第 6 条）。
        </p>
        <p>
          <strong>5.4</strong>{" "}
          您也可以随时清除浏览器中已存储的 Cookie 数据。清除后，下次访问本平台时，必要性
          Cookie 将会重新写入（这是正常登录流程的一部分），而功能性与分析性 Cookie
          将根据您的浏览器设置决定是否写入。
        </p>
      </DocSection>

      {/* clause-6 */}
      <DocSection id="clause-6" title="第6条 本政策的更新">
        <p>
          <strong>6.1</strong>{" "}
          我们可能根据法律法规变化或业务发展需要更新本 Cookie
          政策。重大更新将通过平台公告、邮箱或站内消息通知您，并在政策文本页面醒目标注最近更新日期。
        </p>
        <p>
          <strong>6.2</strong>{" "}
          本政策是 <AgreementLink slug="privacy-policy" />{" "}
          的组成部分；若两者冲突，以 <AgreementLink slug="privacy-policy" />{" "}
          为准。
        </p>
        <p>
          <strong>6.3</strong>{" "}
          您继续使用本平台即视为您接受本政策的更新版本。如您不同意更新内容，可按{" "}
          <AgreementLink slug="privacy-policy" />{" "}
          第 6 条的规定申请注销账号，停止使用本平台服务。
        </p>
      </DocSection>

      {/* clause-7 */}
      <DocSection id="clause-7" title="第7条 联系我们">
        <p>
          如您对本 Cookie 政策有疑问或建议，请通过{" "}
          <P>{PLACEHOLDERS.supportEmail}</P>{" "}
          联系我们，我们将在 10 个工作日内予以回复。
        </p>
        <p>
          若您认为本平台的 Cookie
          使用行为违反了相关法律法规，您有权向国家互联网信息办公室或您所在地的网信部门进行投诉或举报。
        </p>
      </DocSection>
    </>
  );
}
