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

export function PrivacyPolicyDoc() {
  return (
    <>
      <LegalPlaceholderNotice
        fields={["companyName", "dataStorage", "supportEmail"]}
      />
      <p className="text-xs text-slate-500">
        生效日期：{EFFECTIVE_DATE} · 最近更新：{LAST_UPDATED}
      </p>

      {/* Preamble */}
      <p>
        感谢您选择使用{BRAND}（域名：{DOMAIN}
        ）的服务。{BRAND}
        是专注于英语学习的在线服务平台，通过游戏化学习机制、社区功能及
        AI
        辅助能力为用户提供有趣、高效的英语学习体验。{BRAND}运营主体为{" "}
        <LegalPlaceholder>{PLACEHOLDERS.companyName}</LegalPlaceholder>（以下简称
        &ldquo;本平台&rdquo;或&ldquo;运营方&rdquo;）。
      </p>
      <p>
        本隐私政策依据《中华人民共和国个人信息保护法》《中华人民共和国未成年人网络保护条例》《网络信息内容生态治理规定》等法律法规制定，遵循合法、正当、必要、诚信的原则收集和处理您的个人信息。
      </p>
      <p>
        您通过点击&ldquo;同意&rdquo;按钮、注册账号或使用本平台服务，即视为您已充分阅读、理解并接受本隐私政策的全部条款，并同意本平台按本政策的规定处理您的个人信息。
      </p>
      <p>
        若您为未成年人（不满 18
        周岁），请在法定监护人陪同下阅读本隐私政策，并取得监护人明确同意（勾选{" "}
        <AgreementLink slug="guardian-consent" />
        ）后，方可注册或使用本平台服务。
      </p>

      {/* clause-1 */}
      <DocSection id="clause-1" title="第1条 定义">
        <p>在本隐私政策中，下列术语具有以下含义：</p>
        <ol className="list-decimal space-y-3 pl-5">
          <li>
            <strong>个人信息</strong>
            ：指以电子或者其他方式记录的与已识别或者可识别的自然人有关的各种信息，不包括匿名化处理后的信息。本平台收集的个人信息包括但不限于用户名、电子邮箱、学习记录、设备信息、IP
            地址等。
          </li>
          <li>
            <strong>敏感个人信息</strong>
            ：指一旦泄露或者非法使用，容易导致自然人的人格尊严受到侵害或者人身、财产安全受到危害的个人信息。本平台可能处理的敏感个人信息包括：手机号码、身份证号码、未成年人监护人联系方式、未成年人身份信息。对于上述敏感个人信息，本平台将单独取得您的明确同意后方予收集和使用。
          </li>
          <li>
            <strong>匿名化处理</strong>
            ：指对个人信息进行处理，使其无法识别特定自然人且处理后不能复原的过程。经匿名化处理后的信息不属于个人信息，本平台可将其用于统计分析、产品改进等目的，无需另行取得您的同意。
          </li>
          <li>
            <strong>第三方平台</strong>
            ：指本平台在提供服务过程中接入或合作的外部服务提供方，包括但不限于提供微信等快捷登录功能的第三方身份验证服务方，以及提供支付结算功能的第三方支付服务方。本平台与第三方平台之间的数据共享受本政策第4条的约束。
          </li>
        </ol>
      </DocSection>

      {/* clause-2 */}
      <DocSection id="clause-2" title="第2条 个人信息的收集与范围">
        <p>
          本平台遵循最小化收集原则，仅收集为实现服务目的所必要的个人信息。对于敏感个人信息，本平台将在收集前单独向您说明收集目的，并取得您的明确单独同意。
        </p>

        <p>
          <strong>2.1 收集的信息类别</strong>
        </p>
        <table className="w-full border-collapse text-[13px]">
          <thead className="bg-slate-50">
            <tr>
              <th className="border border-slate-200 px-3 py-2 text-left font-semibold">
                信息类别
              </th>
              <th className="border border-slate-200 px-3 py-2 text-left font-semibold">
                具体内容
              </th>
              <th className="border border-slate-200 px-3 py-2 text-left font-semibold">
                收集目的
              </th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td className="border border-slate-200 px-3 py-2 align-top font-medium">
                账号注册与身份验证信息
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                用户名、电子邮箱、微信 openid/unionid、加密存储的密码；未成年人场景下的监护人联系方式
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                创建账号与身份验证；未成年人场景下确认监护人同意
              </td>
            </tr>
            <tr>
              <td className="border border-slate-200 px-3 py-2 align-top font-medium">
                学习相关信息
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                学习时长、课程进度、任务打卡、生词/复习/已掌握记录、排行榜数据、UGC
                内容中的个人信息
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                提供个性化学习推荐；生成学习报告；防止作弊；社区功能使用
              </td>
            </tr>
            <tr>
              <td className="border border-slate-200 px-3 py-2 align-top font-medium">
                第三方授权信息
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                微信快捷登录时的公开信息（昵称、头像）与
                openid/unionid；不收集第三方账号密码、好友列表
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                实现快捷登录，仅用于身份标识
              </td>
            </tr>
            <tr>
              <td className="border border-slate-200 px-3 py-2 align-top font-medium">
                支付相关信息
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                订单号、支付金额、支付时间、支付渠道名称；不收集银行卡号、支付密码
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                确认付费服务订单状态，开通付费权益
              </td>
            </tr>
            <tr>
              <td className="border border-slate-200 px-3 py-2 align-top font-medium">
                设备与访问信息
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                设备型号、操作系统版本、浏览器类型、IP
                地址、登录时间与地点、页面访问路径
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                保障账号安全、适配功能、排查服务故障
              </td>
            </tr>
          </tbody>
        </table>

        <p>
          <strong>2.2 收集方式</strong>
        </p>
        <ul className="list-disc space-y-2 pl-5">
          <li>
            <strong>您主动提供：</strong>
            注册账号、填写个人资料、发布内容、发起客服咨询时，您主动向本平台提供相关信息。
          </li>
          <li>
            <strong>服务使用中自动记录：</strong>
            您使用本平台服务过程中，系统自动收集设备信息、访问日志、学习行为数据等技术信息。
          </li>
          <li>
            <strong>第三方授权获取：</strong>
            您选择使用第三方快捷登录时，本平台依据您的授权从第三方平台获取必要的身份标识信息。
          </li>
        </ul>

        <p>
          <strong>2.3 不收集的信息</strong>
        </p>
        <p>
          本平台不收集与英语学习服务无直接关联的个人信息，包括但不限于宗教信仰、医疗健康记录、金融账户余额、行踪轨迹等。除非法律另有要求，或您在使用服务过程中主动提供上述信息（即便如此，本平台仍会向您提示该信息属于非必要信息，并询问您是否确认提供）。
        </p>
      </DocSection>

      {/* clause-3 */}
      <DocSection id="clause-3" title="第3条 个人信息的使用规则">
        <p>
          本平台对个人信息的使用严格限定于实现服务目的所必要的范围，不会将您的个人信息用于与收集目的无关的其他用途，除非另行取得您的明确同意。
        </p>

        <p>
          <strong>3.1 核心使用场景</strong>
        </p>
        <ol className="list-decimal space-y-3 pl-5">
          <li>
            <strong>3.1.1 提供基础学习服务：</strong>
            使用您的账号信息进行登录身份验证；根据您的学习记录和偏好提供个性化课程推荐；保障平台各项功能的正常运行。
          </li>
          <li>
            <strong>3.1.2 优化服务体验：</strong>
            根据您的学习行为数据（如学习时长、错误率、完成度）动态调整课程难度或内容推荐，提升学习效果和平台使用体验。
          </li>
          <li>
            <strong>3.1.3 账号与交易安全保障：</strong>
            使用验证码验证确保敏感操作安全；核对支付订单信息确保交易准确性；通过设备与访问信息识别和处理异常登录行为。
          </li>
          <li>
            <strong>3.1.4 发送通知：</strong>
            本平台向您发送的通知分为两类：
            <ul className="mt-2 list-disc space-y-1 pl-5">
              <li>
                <strong>必选通知（无法关闭）：</strong>
                验证码、账号安全提醒、付费到期提醒、服务重大变更通知。此类通知属于服务履约所必需，无法通过设置关闭。
              </li>
              <li>
                <strong>可选通知（可关闭）：</strong>
                学习提醒、新课程上线通知、活动优惠信息。您可在&ldquo;个人主页
                — 隐私设置&rdquo;中关闭此类通知。
              </li>
            </ul>
            <p className="mt-2">
              <strong>
                您注册并使用本平台服务，即表示您已阅读、理解并同意我们按上述方式向您发送通知。
              </strong>
            </p>
          </li>
        </ol>

        <p>
          <strong>3.2 合规与争议处理</strong>
        </p>
        <p>
          在司法机关依法要求配合调查，或处理用户之间、用户与平台之间的服务纠纷时，本平台可能使用相关个人信息，但仅限于所必需的最小范围，且严格按照法律规定的程序执行。
        </p>

        <p>
          <strong>3.3 禁止性使用</strong>
        </p>
        <ul className="list-disc space-y-2 pl-5">
          <li>
            本平台不进行&ldquo;大数据杀熟&rdquo;，不针对不同用户群体设置差异化价格或进行价格歧视；
          </li>
          <li>
            本平台不将您的个人信息用于向您推送与本平台服务无关的第三方营销推广内容，除非事先单独取得您的明确同意；
          </li>
          <li>
            本平台不将未成年人的学习记录、身份信息及监护人信息用于任何商业目的，包括但不限于广告投放、数据分析变现等。
          </li>
        </ul>
      </DocSection>

      {/* clause-4 */}
      <DocSection id="clause-4" title="第4条 个人信息的共享、转让与披露">
        <p>
          除法律法规另有规定，或经您书面明确同意外，本平台不向任何第三方共享、转让或披露您的个人信息。
        </p>

        <p>
          <strong>4.1 共享</strong>
        </p>
        <p>
          本平台仅在以下有限场景中，与经严格审核的合作方共享必要的个人信息：
        </p>
        <ul className="list-disc space-y-2 pl-5">
          <li>
            <strong>（1）支付合作方：</strong>
            为确认付费服务订单状态，仅向支付服务方共享订单号、支付金额、注册手机号等订单核对信息；不共享银行卡号、支付密码等敏感支付信息。
          </li>
          <li>
            <strong>（2）课程内容合作方：</strong>
            为评估课程内容效果及提供学习报告，可能向课程内容合作方共享匿名化或脱敏后的学习进度数据；不共享可识别您身份的个人信息。
          </li>
        </ul>
        <p>上述共享均受到以下约束：</p>
        <ul className="list-disc space-y-1 pl-5">
          <li>
            （a）合作方仅可在约定的目的范围内使用共享信息，不得超出授权范围另行处理；
          </li>
          <li>
            （b）合作方须采取与本平台同等或更高级别的安全保护措施；
          </li>
          <li>
            （c）服务合作关系结束后，合作方须及时删除或依约返还相关信息。
          </li>
        </ul>

        <p>
          <strong>4.2 转让</strong>
        </p>
        <p>
          本平台不会将您的个人信息转让给第三方，但以下业务重组场景除外：合并、收购、资产转让等情形下，如涉及个人信息的转移，本平台将：
        </p>
        <ul className="list-disc space-y-2 pl-5">
          <li>在转让发生前至少 7 日通过平台公告方式向您告知；</li>
          <li>
            要求接收方承诺继续履行本隐私政策的全部义务，并受本政策约束；
          </li>
          <li>
            如转让涉及敏感个人信息，将在公告之外额外取得您的单独书面同意。
          </li>
        </ul>

        <p>
          <strong>4.3 披露</strong>
        </p>
        <p>仅在以下法定情形下，本平台方可披露您的个人信息：</p>
        <ul className="list-disc space-y-2 pl-5">
          <li>
            依据人民法院、检察院、公安机关等有权机关依法出具的合法文书要求；
          </li>
          <li>
            为保护您或其他用户的生命、财产等重大合法权益，且无法取得当事人同意的紧急情形；
          </li>
          <li>法律法规明确规定的其他情形，如维护国家安全或公共利益所必需。</li>
        </ul>

        <DocCallout variant="info" title="信息共享提示">
          <p>
            本平台不出售您的个人信息。所有共享行为均以保障您的服务体验为前提，并受到严格的合同约束和安全审查。
          </p>
        </DocCallout>
      </DocSection>

      {/* clause-5 */}
      <DocSection id="clause-5" title="第5条 个人信息的存储与安全保护">
        <p>
          <strong>5.1 存储规则</strong>
        </p>

        <p>
          <strong>5.1.1 存储地点</strong>
        </p>
        <p>
          您的个人信息存储于位于{" "}
          <LegalPlaceholder>{PLACEHOLDERS.dataStorage}</LegalPlaceholder>{" "}
          的境内服务器，本平台不对您的个人信息进行跨境传输或存储，法律法规另有规定的除外。
        </p>

        <p>
          <strong>5.1.2 存储期限</strong>
        </p>
        <p>本平台在以下期限内保存您的个人信息：</p>
        <ul className="list-disc space-y-2 pl-5">
          <li>
            账号存续期间：账号处于正常状态期间，本平台持续保存您的账号信息及学习记录；
          </li>
          <li>
            账号注销后：您注销账号后，本平台将在 30
            个工作日内删除或匿名化处理您的个人信息；
          </li>
          <li>
            法定保留期限例外：法律法规要求保留的信息（如支付记录须保存
            5 年）按法定期限执行，届满后予以删除或匿名化。
          </li>
        </ul>

        <p>
          <strong>5.1.3 存储形式</strong>
        </p>
        <p>
          密码等敏感账号信息采用不可逆加密方式存储；其他个人信息采用加密传输和加密存储方式保护，防止未经授权的访问和泄露。
        </p>

        <p>
          <strong>5.2 安全保护措施</strong>
        </p>

        <p>
          <strong>5.2.1 技术措施</strong>
        </p>
        <ul className="list-disc space-y-2 pl-5">
          <li>
            <strong>访问控制：</strong>
            仅向工作职责有必要的员工授予访问个人信息的权限，并实施最小权限原则；
          </li>
          <li>
            <strong>安全监测：</strong>
            对数据访问行为实施持续监测，批量下载或异常访问行为将自动触发安全报警；
          </li>
          <li>
            <strong>数据备份：</strong>
            每日对重要数据进行异地备份，保障数据可恢复性。
          </li>
        </ul>

        <p>
          <strong>5.2.2 管理措施</strong>
        </p>
        <ul className="list-disc space-y-2 pl-5">
          <li>
            对全体员工开展个人信息保护及安全意识培训；
          </li>
          <li>
            与所有可能接触用户个人信息的员工及外部合作方签署《保密协议》；
          </li>
          <li>
            建立个人信息安全事件应急响应预案，发生安全事件后 1
            日内启动应急响应，并按法律规定向监管部门报告。
          </li>
        </ul>

        <DocCallout variant="warning" title="用户侧安全提示">
          <p>
            为保障您的账号安全，请妥善保管账号密码，避免在公共设备上登录本平台，定期更新密码，并为账号绑定安全的电子邮箱。如发现账号存在异常登录或未授权操作，请立即通过&ldquo;个人主页
            — 账号安全&rdquo;冻结账号，并联系 <LegalPlaceholder>{PLACEHOLDERS.supportEmail}</LegalPlaceholder>。
          </p>
        </DocCallout>
      </DocSection>

      {/* clause-6 */}
      <DocSection id="clause-6" title="第6条 用户的个人信息权利及行使方式">
        <p>
          依据《中华人民共和国个人信息保护法》及相关法律法规，您对本平台处理的个人信息依法享有查询、更正、删除、撤回同意、注销账号等权利。本平台将依法保障您权利的有效行使。
        </p>

        <p>
          <strong>6.1 权利行使方式</strong>
        </p>
        <table className="w-full border-collapse text-[13px]">
          <thead className="bg-slate-50">
            <tr>
              <th className="border border-slate-200 px-3 py-2 text-left font-semibold">
                权利类型
              </th>
              <th className="border border-slate-200 px-3 py-2 text-left font-semibold">
                具体内容
              </th>
              <th className="border border-slate-200 px-3 py-2 text-left font-semibold">
                行使路径
              </th>
              <th className="border border-slate-200 px-3 py-2 text-left font-semibold">
                响应时间
              </th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td className="border border-slate-200 px-3 py-2 align-top font-medium">
                查询权
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                查阅账号基础信息、学习记录及付费记录
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                个人主页 → 基础信息 / 学习记录
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                即时
              </td>
            </tr>
            <tr>
              <td className="border border-slate-200 px-3 py-2 align-top font-medium">
                更正权
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                修改错误或过时的个人信息（如邮箱变更、昵称修改）
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                （a）自主更正：个人主页 → 编辑 → 个人资料 / 账号与安全；（b）客服协助更正：发送邮件至{" "}
                <LegalPlaceholder>{PLACEHOLDERS.supportEmail}</LegalPlaceholder>
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                自主即时 / 客服 10 个工作日内
              </td>
            </tr>
            <tr>
              <td className="border border-slate-200 px-3 py-2 align-top font-medium">
                删除权
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                删除非必要的个人信息及发布的 UGC 内容
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                （a）自主删除：个人主页 → 注销账号；（b）客服协助：发送邮件至{" "}
                <LegalPlaceholder>{PLACEHOLDERS.supportEmail}</LegalPlaceholder>
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                自主即时 / 客服 10 个工作日内
              </td>
            </tr>
            <tr>
              <td className="border border-slate-200 px-3 py-2 align-top font-medium">
                撤回同意权
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                撤回对非必要个人信息收集的同意，关闭个性化推荐及可选通知
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                个人主页 → 隐私设置 → 关闭相关授权
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                即时
              </td>
            </tr>
            <tr>
              <td className="border border-slate-200 px-3 py-2 align-top font-medium">
                注销权
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                申请注销账号，终止本平台对您个人信息的持续处理
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                个人主页 → 账号安全 → 注销
              </td>
              <td className="border border-slate-200 px-3 py-2 align-top">
                即时（完成注销流程后）
              </td>
            </tr>
          </tbody>
        </table>

        <p>
          <strong>6.2 说明</strong>
        </p>
        <ul className="list-disc space-y-2 pl-5">
          <li>
            <strong>身份验证：</strong>
            为保障账号安全，您在行使上述权利前，本平台将通过向您注册的电子邮箱发送验证码的方式验证您的身份，验证通过后方可执行相关操作。
          </li>
          <li>
            <strong>功能影响：</strong>
            撤回同意或删除部分信息可能导致相关服务功能无法正常使用（例如，关闭学习行为记录可能影响个性化课程推荐功能），本平台将在操作前向您说明可能的影响。
          </li>
          <li>
            <strong>投诉渠道：</strong>
            若您认为本平台的个人信息处理行为违反法律法规，可向国家互联网信息办公室或其他有权监管部门进行投诉举报。
          </li>
        </ul>
      </DocSection>

      {/* clause-7 */}
      <DocSection id="clause-7" title="第7条 未成年人个人信息的特别保护">
        <p>
          本平台高度重视对未成年人个人信息的保护，并依据《中华人民共和国未成年人网络保护条例》等法律法规对未成年用户实施专项保护措施。
        </p>

        <p>
          <strong>7.1 未成年人账号注册限制</strong>
        </p>
        <ul className="list-disc space-y-2 pl-5">
          <li>
            <strong>8 周岁以下：</strong>
            不得注册本平台账号。若发现此类误注册账号，监护人可通过{" "}
            <LegalPlaceholder>{PLACEHOLDERS.supportEmail}</LegalPlaceholder>{" "}
            反馈，本平台将在 5 个工作日内注销账号并删除相关信息。
          </li>
          <li>
            <strong>8 周岁以上不满 16 周岁：</strong>
            须由法定监护人协助完成注册，并由监护人明确勾选确认{" "}
            <AgreementLink slug="guardian-consent" />
            ；缺少监护人有效同意的注册无效。使用任何付费功能须经监护人明确同意。
          </li>
          <li>
            <strong>16 周岁以上不满 18 周岁：</strong>
            可独立完成基础注册，但使用任何付费功能仍须取得监护人明确同意（勾选{" "}
            <AgreementLink slug="guardian-consent" />）。
          </li>
        </ul>

        <p>
          <strong>7.2 监护人权利</strong>
        </p>
        <p>
          监护人可通过 <LegalPlaceholder>{PLACEHOLDERS.supportEmail}</LegalPlaceholder>{" "}
          客服渠道对未成年人的个人信息行使以下权利：
        </p>
        <ul className="list-disc space-y-2 pl-5">
          <li>
            <strong>查询权：</strong>
            查询未成年人的学习记录、付费记录及账号信息；
          </li>
          <li>
            <strong>删除权：</strong>
            申请删除未成年人发布的 UGC 内容及学习记录；
          </li>
          <li>
            <strong>异议权：</strong>
            对本平台处理未成年人个人信息的行为提出异议，要求本平台说明处理依据，并可要求本平台调整或停止相关处理行为。
          </li>
        </ul>

        <p>
          <strong>7.3 信息处理限制</strong>
        </p>
        <ul className="list-disc space-y-2 pl-5">
          <li>
            本平台不向未成年用户推送商业广告或与英语学习无关的内容推荐；
          </li>
          <li>
            本平台不收集未成年人的生物识别信息（如人脸识别、指纹）、行踪轨迹信息；
          </li>
          <li>
            未成年人的学习记录、身份信息及监护人联系信息不用于任何商业目的，不向第三方共享（服务所必需且经监护人同意的除外）；
          </li>
          <li>
            本平台依法对未成年用户实施使用时长管理及消费限额限制，具体规则以平台公告为准。
          </li>
        </ul>
      </DocSection>

      {/* clause-8 */}
      <DocSection id="clause-8" title="第8条 隐私政策的修改与通知">
        <p>
          <strong>8.1 修改规则</strong>
        </p>
        <p>
          本平台可能根据法律法规的变化、监管要求的更新，或为改进服务体验，对本隐私政策进行修订和更新。修订后的政策将在本平台内公告，在遵守适用法律的前提下，修订生效后继续使用本平台服务视为接受修订内容。
        </p>

        <p>
          <strong>8.2 修改通知与生效</strong>
        </p>
        <ul className="list-disc space-y-2 pl-5">
          <li>
            <strong>通知方式：</strong>
            修订后的隐私政策将在&ldquo;个人主页 →
            隐私设置&rdquo;板块发布公告，并在政策文本页面醒目标注最近更新日期，方便您追踪变更。
          </li>
          <li>
            <strong>生效规则：</strong>
            政策修订公告发出后，在异议期内未提出异议并继续使用本平台服务，视为接受修订后的政策。若您不同意修订内容，可按本政策第6条的规定申请注销账号，停止使用本平台服务。
          </li>
        </ul>
      </DocSection>

      {/* clause-9 */}
      <DocSection id="clause-9" title="第9条 联系我们">
        <p>
          如您对本隐私政策有任何疑问、意见或建议，或需要行使本政策项下的个人信息权利，请通过以下方式联系本平台：
        </p>
        <ul className="list-disc space-y-2 pl-5">
          <li>
            <strong>电子邮箱：</strong>
            <LegalPlaceholder>{PLACEHOLDERS.supportEmail}</LegalPlaceholder>
          </li>
        </ul>
        <p>
          本平台将在收到您的请求后 10
          个工作日内予以响应。如因信息提供不完整或身份验证未通过，本平台可能需要与您进一步沟通确认，响应时间自信息补全之日起计算。
        </p>
        <p>
          若您对本平台的处理结果不满意，或认为本平台的个人信息处理行为侵害了您的合法权益，您有权向国家互联网信息办公室或您所在地的网信部门、公安机关等有权监管机构投诉或举报。
        </p>
      </DocSection>

      {/* clause-10 */}
      <DocSection id="clause-10" title="第10条 其他">
        <p>
          10.1 本隐私政策与 <AgreementLink slug="user-agreement" />{" "}
          具有同等法律效力，共同构成用户与本平台之间关于个人信息处理事项的完整约定。若本隐私政策与{" "}
          <AgreementLink slug="user-agreement" />{" "}
          的条款存在冲突，<strong>以本隐私政策为准</strong>
          （优先保护用户的个人信息权益）。
        </p>
        <p>
          10.2 本隐私政策各条款相互独立。如任何条款被有权机关认定为无效或不可执行，不影响其他条款的合法性和效力；无效或不可执行的条款将被修改为在法律允许的最大范围内与原条款意图最接近的有效条款。
        </p>
        <p>
          在适用法律允许的范围内，本隐私政策的最终解释权归{" "}
          <LegalPlaceholder>{PLACEHOLDERS.companyName}</LegalPlaceholder>{" "}
          所有。如对本政策有任何疑问，请通过第9条所列联系方式与我们取得联系。
        </p>
      </DocSection>
    </>
  );
}
