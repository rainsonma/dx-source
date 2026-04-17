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

export function ProductServiceDoc() {
  return (
    <>
      <LegalPlaceholderNotice
        fields={["companyName", "companyAddr", "courtLocation", "supportEmail"]}
      />
      <p className="text-xs text-slate-500">
        生效日期：{EFFECTIVE_DATE} · 最近更新：{LAST_UPDATED}
      </p>

      {/* Preamble */}
      <p>
        欢迎您订阅{BRAND}（域名：{DOMAIN}
        ）提供的会员服务。{BRAND}
        运营主体为 <LegalPlaceholder>{PLACEHOLDERS.companyName}</LegalPlaceholder>（注册地址：
        <LegalPlaceholder>{PLACEHOLDERS.companyAddr}</LegalPlaceholder>，以下简称 &ldquo;本平台&rdquo;
        或 &ldquo;运营方&rdquo;）。
      </p>
      <p>
        本平台通过 AI 技术赋能课程、游戏化学习机制（任务、打卡、积分、排行榜）及社区
        / 小组功能，为会员提供优质、高效的英语学习体验。
      </p>
      <p>
        点击 &ldquo;同意订阅&rdquo; / &ldquo;确认支付&rdquo; 按钮、注册或使用会员服务，即视为您已充分阅读、理解并完全接受本协议全部条款（
        <strong>
          特别是加粗标注的免除 / 限制本平台责任、加重您义务、涉及您重大权益的条款
        </strong>
        ）。
      </p>
      <p>
        若您为未成年人，须在监护人同意（勾选{" "}
        <AgreementLink slug="guardian-consent" />
        ）后方可订阅本会员服务。
      </p>
      <p>
        本协议依据《中华人民共和国网络安全法》《中华人民共和国个人信息保护法》《中华人民共和国消费者权益保护法》《网络信息内容生态治理规定》等法律法规制定。
      </p>

      {/* clause-1 */}
      <DocSection id="clause-1" title="第1条 定义">
        <p>在本协议中，下列术语具有以下含义：</p>
        <ol className="list-decimal space-y-3 pl-5">
          <li>
            <strong>用户 / 会员</strong>
            ：指完成注册并订阅本平台会员服务、在会员周期内享有会员权益的自然人。
          </li>
          <li>
            <strong>本平台</strong>
            ：指由 <LegalPlaceholder>{PLACEHOLDERS.companyName}</LegalPlaceholder> 运营的{BRAND}产品及其相关服务的统称，涵盖网站（域名：{DOMAIN}）、移动应用、小程序等所有形态的终端。
          </li>
          <li>
            <strong>会员服务</strong>
            ：指本平台向订阅用户提供的标注 &ldquo;会员专享&rdquo; / &ldquo;会员共享&rdquo; 的付费服务，包括但不限于课程解锁、虚拟物品权益、社区特权及功能权益。
          </li>
          <li>
            <strong>会员周期</strong>
            ：指用户订阅的会员服务有效期，分为月度、季度、年度及终身四种档位。
          </li>
          <li>
            <strong>会员费用</strong>
            ：指用户为获取会员服务须支付的对价，具体金额以订阅页面实时展示为准。
          </li>
          <li>
            <strong>虚拟物品</strong>
            ：指本平台发行的、仅限在本平台内使用的虚拟权益凭证，包括<strong>能量豆</strong>等。能量豆不具备法定货币属性，不可兑换现金，亦不可转让给他人。
          </li>
          <li>
            <strong>游戏化作弊行为</strong>
            ：指使用外挂、脚本、刷分工具、批量账号或任何技术手段干扰任务积分、等级、排行榜数据真实性的行为。
          </li>
        </ol>
      </DocSection>

      {/* clause-2 */}
      <DocSection id="clause-2" title="第2条 协议生效与效力">
        <p>
          <strong>2.1 生效触发条件</strong>
        </p>
        <p>下列任一行为发生，本协议即对您生效：</p>
        <ol className="list-decimal space-y-3 pl-5">
          <li>
            点击 &ldquo;同意协议并支付&rdquo; 按钮；
          </li>
          <li>
            <strong>
              付费行为：您完成会员费用支付，无论是否点击协议确认按钮，付费行为本身即视为您已充分阅读、理解并完全接受本协议全部条款；
            </strong>
          </li>
          <li>
            实际使用行为：您在付费后实际使用会员服务，即视为您认可本协议效力。
          </li>
        </ol>

        <p>
          <strong>2.2 效力范围</strong>
        </p>
        <ul className="list-disc space-y-2 pl-5">
          <li>
            <strong>对用户的效力：</strong>
            本协议对订阅用户具有法律约束力，您须遵守本协议的全部条款。
          </li>
          <li>
            <strong>对平台的效力：</strong>
            本平台承诺依据本协议向您提供会员服务，并在法律范围内保障您的合法权益。
          </li>
          <li>
            <strong>协议冲突处理：</strong>
            本协议与 <AgreementLink slug="user-agreement" />、
            <AgreementLink slug="privacy-policy" />{" "}
            存在冲突时，以本协议为准；本协议未约定的事项，适用上述其他协议。
          </li>
        </ul>

        <p>
          <strong>2.3 条款独立性</strong>
        </p>
        <p>
          本协议各条款相互独立。如任何条款被有权机关认定为无效或不可执行，不影响其他条款的合法性和效力。
        </p>
      </DocSection>

      {/* clause-3 */}
      <DocSection id="clause-3" title="第3条 会员服务内容与权益">
        <p>
          <strong>3.1 核心会员权益</strong>
        </p>
        <ol className="list-decimal space-y-3 pl-5">
          <li>
            <strong>课程权益：</strong>
            解锁所有标注 &ldquo;会员专享&rdquo;、&ldquo;会员共享&rdquo; 的英文课程与关卡。
          </li>
          <li>
            <strong>功能权益：</strong>
            学习数据可视化、会员专属客服通道。
          </li>
          <li>
            <strong>虚拟物品权益：</strong>
            能量豆月度赠送，排行榜专属标识。
          </li>
          <li>
            <strong>社区权益：</strong>
            会员专属学习社群 / 小组创建权益。
          </li>
          <li>
            <strong>推广返利权益：</strong>
            推广返利的具体规则以平台《推广规则》实时公示为准。
          </li>
        </ol>

        <p>
          <strong>3.2 权益限制</strong>
        </p>
        <ol className="list-decimal space-y-3 pl-5">
          <li>
            <strong>会员权益仅在会员周期内有效，到期未续费则权益自动终止；</strong>
          </li>
          <li>
            <strong>会员权益仅限订阅账号本人使用，不可转让、出借、出租或与他人共享；</strong>
          </li>
          <li>
            平台可基于技术升级、课程迭代调整权益（提前 1 日公告），但不削减核心权益。
          </li>
        </ol>
      </DocSection>

      {/* clause-4 */}
      <DocSection id="clause-4" title="第4条 会员订阅与支付规则">
        <p>
          <strong>4.1 订阅流程</strong>
        </p>
        <ol className="list-decimal space-y-3 pl-5">
          <li>注册登录并绑定邮箱；</li>
          <li>选择订阅档位（月度 / 季度 / 年度 / 终身会员）；</li>
          <li>确认订单信息（会员周期、费用）；</li>
          <li>通过平台支持的付费渠道完成支付。</li>
        </ol>

        <p>
          <strong>4.2 费用说明</strong>
        </p>
        <ul className="list-disc space-y-2 pl-5">
          <li>会员费用为含税价格；</li>
          <li>
            本平台可能推出会员优惠活动，优惠规则以活动页面说明为准。
          </li>
        </ul>
      </DocSection>

      {/* clause-5 */}
      <DocSection id="clause-5" title="第5条 会员账号使用规范">
        <p>
          <strong>5.1 账号归属</strong>
        </p>
        <DocCallout variant="warning" title="账号归属与禁止共享">
          <p>
            会员账号归本平台所有，您仅享有会员服务的使用权。
          </p>
          <p>
            <strong>
              不得将账号转借、出租、出售或共享给他人（如多人共用一个会员账号学习）。
            </strong>
          </p>
          <p>
            发现违规行为，本平台有权暂停会员权益、封禁账号，且不退还已支付的会员费用。
          </p>
        </DocCallout>

        <p>
          <strong>5.2 账号保管</strong>
        </p>
        <p>
          妥善保管账号密码；因自身原因导致账号被盗的，本平台不承担责任。发现账号异常请立即联系客服冻结账号。
        </p>

        <p>
          <strong>5.3 单设备登录</strong>
        </p>
        <p>
          本平台实施账号同一时间仅限单设备登录的技术限制措施。若检测到异常登录行为，本平台有权要求身份验证或临时冻结账号。
        </p>

        <p>
          <strong>5.4 未成年人使用</strong>
        </p>
        <ol className="list-decimal space-y-3 pl-5">
          <li>不满 8 周岁的未成年人不得订阅会员服务；</li>
          <li>
            8 周岁以上不满 16 周岁的未成年人，需由监护人协助订阅并勾选{" "}
            <AgreementLink slug="guardian-consent" />；
          </li>
          <li>
            16 周岁以上不满 18 周岁的未成年人，需经监护人同意（勾选{" "}
            <AgreementLink slug="guardian-consent" />
            ）后方可订阅。
          </li>
        </ol>
      </DocSection>

      {/* clause-6 */}
      <DocSection id="clause-6" title="第6条 会员服务的暂停与终止">
        <p>
          <strong>6.1 暂停</strong>
        </p>
        <ol className="list-decimal space-y-3 pl-5">
          <li>
            因本平台技术升级、系统维护需暂停会员服务的，将提前 1 日通过公告通知用户；若单次暂停达到或超过 24 小时，会员有效期将按实际暂停时长对等顺延；
          </li>
          <li>
            若您违反本协议约定，本平台有权暂停会员权益 3–7 日，暂停期间不延长会员周期，且不退还会员费用。
          </li>
        </ol>

        <p>
          <strong>6.2 终止</strong>
        </p>
        <ol className="list-decimal space-y-3 pl-5">
          <li>会员周期到期且未续费的，会员服务自动终止；</li>
          <li>
            您主动申请注销账号的，会员服务自账号注销完成之日起终止，未到期的会员权益不予退款；
          </li>
          <li>
            <strong>
              您严重违反本协议（如出租、出售会员账号、盗用他人账号订阅），本平台有权单方面终止会员服务、封禁账号，且不退还已支付的会员费用；
            </strong>
          </li>
          <li>
            本平台因业务调整停止会员服务的，将提前 30 日发布公告，为未到期会员按 &ldquo;未使用天数 / 总天数 &times; 已支付费用&rdquo; 办理退款。
          </li>
        </ol>
      </DocSection>

      {/* clause-7 */}
      <DocSection id="clause-7" title="第7条 退订与退款规则">
        <p>
          <strong>7.1 生效条款</strong>
        </p>
        <p>
          您确认，本协议第 2 条 &ldquo;协议生效与效力&rdquo; 中 &ldquo;付费即视为同意协议&rdquo; 的约定已包含您对本退款规则的认可。
        </p>

        <p>
          <strong>7.2 数字商品不可退款说明</strong>
        </p>
        <p>
          会员服务属于《中华人民共和国消费者权益保护法》第二十五条规定的 &ldquo;在线下载的数字化商品&rdquo;，具有无形性、即时交付性、不可回收性等特点，除本协议明确约定外，
          <strong>订阅后不支持无理由退款</strong>。
        </p>
        <p>订阅前，本平台通过以下方式履行告知义务：</p>
        <ol className="list-decimal space-y-3 pl-5">
          <li>
            <strong>显著位置提示：</strong>
            在会员订阅页面顶部、支付确认页以加粗文字标注 &ldquo;虚拟商品，订阅后不支持无理由退款，请确认后支付&rdquo;；
          </li>
          <li>
            <strong>弹窗二次确认：</strong>
            支付前弹出独立窗口，要求您勾选 &ldquo;已阅读并同意{" "}
            <AgreementLink slug="product-service" />，确认虚拟商品不支持无理由退款&rdquo;；
          </li>
          <li>
            <strong>协议条款加粗：</strong>
            本协议第 7 条以加粗字体明确不可退款规则。
          </li>
        </ol>

        <p>
          <strong>7.3 例外可退款情形（仅以下情况支持退款）</strong>
        </p>
        <ol className="list-decimal space-y-3 pl-5">
          <li>
            <strong>年度会员 / 终身会员 10 天无理由退款：</strong>
            年度会员或终身会员开通后 10 个自然日内（自付款时间起算），用户可联系客服申请无理由全额退款。退款申请一经核实通过，会员服务立即终止。
            <strong>本条款不适用于月度会员、季度会员及能量豆充值。</strong>
          </li>
          <li>
            <strong>未成年人误订阅：</strong>
            8 周岁以上未成年人未经监护人同意订阅会员服务，监护人可凭户口本 / 出生证明等监护关系证明申请全额退款（需在订阅后 7 日内提交，且未成年人账号未使用任何会员权益）。
          </li>
          <li>
            <strong>系统故障误扣：</strong>
            因系统故障导致重复扣费、错扣费用，经核实后 100% 退还。
          </li>
        </ol>

        <p>
          <strong>7.4 不可退款的明确情形</strong>
        </p>
        <ol className="list-decimal space-y-3 pl-5">
          <li>月度会员、季度会员一经购买不支持退款（年度 / 终身另适用 7.3(1)）；</li>
          <li>年度 / 终身会员超过 10 个自然日的退款申请；</li>
          <li>因您自身原因（如更换学习目标、误操作订阅）在退款窗口外申请；</li>
          <li>违反本协议导致账号被封禁（如作弊、转让账号）。</li>
        </ol>

        <p>
          <strong>7.5 退款流程与时效</strong>
        </p>
        <ol className="list-decimal space-y-3 pl-5">
          <li>
            <strong>申请路径：</strong>
            通过 <LegalPlaceholder>{PLACEHOLDERS.supportEmail}</LegalPlaceholder> 提交退款申请，并按客服指引提供材料；
          </li>
          <li>
            <strong>审核时效：</strong>
            10 个工作日内反馈；
          </li>
          <li>
            <strong>退款方式：</strong>
            原路返回至您的支付账户。
          </li>
        </ol>

        <p>
          <strong>7.6 未成年人退款特别规则</strong>
        </p>
        <ul className="list-disc space-y-2 pl-5">
          <li>
            未成年人使用过会员权益的，仅支持退还未使用部分费用（按 &ldquo;剩余会员天数 / 总天数 &times; 订阅费用&rdquo; 计算）；
          </li>
          <li>
            监护人须在未成年人首次订阅后 7 日内提出退款，逾期视为同意订阅（以订单支付时间为准）。
          </li>
        </ul>
      </DocSection>

      {/* clause-8 */}
      <DocSection id="clause-8" title="第8条 知识产权">
        <p>
          <strong>8.1 本平台知识产权</strong>
        </p>
        <p>
          课程内容、AI 算法模型、软件代码、数据库结构、界面设计、会员虚拟道具设计、游戏化学习机制（积分、排行榜、任务体系）、商标、Logo、域名均归本平台所有，受著作权法、商标法等知识产权法律保护。
        </p>

        <p>
          <strong>8.2 用户禁止行为</strong>
        </p>
        <ul className="list-disc space-y-2 pl-5">
          <li>不得复制、下载、录屏传播课程内容；</li>
          <li>不得篡改、拆解算法模型、反编译软件；</li>
          <li>不得盗用商标 / Logo / 界面设计；</li>
          <li>未经书面授权不得将本平台知识产权用于商业目的。</li>
        </ul>

        <p>
          <strong>8.3 UGC 知识产权</strong>
        </p>
        <p>
          用户 UGC 内容著作权归用户所有；用户授予本平台全球范围内、非独占、免费、可转授权的许可（在本平台内展示、传播、服务优化、合法推广）。用户保证 UGC 不侵犯第三方合法权益；因 UGC 侵权导致本平台损失的，用户须承担全部责任。
        </p>

        <p>
          <strong>8.4</strong>{" "}
          本条款未明确授予的知识产权相关权利均视为未授权，不得自行主张或使用。
        </p>
      </DocSection>

      {/* clause-9 */}
      <DocSection id="clause-9" title="第9条 免责声明">
        <p>
          <strong>9.1 学习效果免责</strong>
        </p>
        <p>
          本平台提供英文课程及学习工具，
          <strong>
            学习效果受您个人基础、学习时长、努力程度等因素影响，本平台不对学习结果作出任何明示或暗示的保证。
          </strong>
        </p>

        <p>
          <strong>9.2 服务中断免责</strong>
        </p>
        <p>
          <strong>
            因不可抗力（自然灾害、政策调整）、第三方原因（网络运营商故障、支付渠道问题）导致服务中断、数据丢失的，本平台将尽合理努力恢复服务，但不承担间接损失。
          </strong>
        </p>

        <p>
          <strong>9.3 技术限制免责</strong>
        </p>
        <p>
          AI 功能受技术水平限制，本平台持续优化，但不保证完全无误差。
        </p>
      </DocSection>

      {/* clause-10 */}
      <DocSection id="clause-10" title="第10条 账号注销">
        <p>
          <strong>10.1 注销条件</strong>
        </p>
        <ul className="list-disc space-y-2 pl-5">
          <li>账号处于正常状态（无违规封禁、无纠纷）；</li>
          <li>账号内无未使用的虚拟物品或未到期权益（或已书面放弃，注销后不予退款）。</li>
        </ul>

        <p>
          <strong>10.2 注销流程</strong>
        </p>
        <ul className="list-disc space-y-2 pl-5">
          <li>
            <strong>（a）自主提交路径：</strong>
            登录账号 → 账号安全 → 选择注销账号 → 完成身份验证 → 确认注销须知并提交申请；
          </li>
          <li>
            <strong>（b）协助提交路径：</strong>
            通过 <LegalPlaceholder>{PLACEHOLDERS.supportEmail}</LegalPlaceholder> 向客服提交注销申请。
          </li>
        </ul>

        <p>
          <strong>10.3 注销后果</strong>
        </p>
        <p>
          <strong>
            账号注销后无法登录或使用该账号，账号内的用户权益、虚拟物品（含能量豆）、学习记录、社区数据等将永久删除且不可恢复。
          </strong>
        </p>
      </DocSection>

      {/* clause-11 */}
      <DocSection id="clause-11" title="第11条 协议的修改与通知">
        <p>
          <strong>11.1 修改</strong>
        </p>
        <p>
          平台有权根据法律法规或运营需求修改本协议，修改后的协议将在平台内公告并标注 &ldquo;更新日期&rdquo;；您继续使用会员服务视为接受；不同意可申请注销。
        </p>

        <p>
          <strong>11.2 通知送达</strong>
        </p>
        <p>
          平台通知可通过公告、注册邮箱推送，视为有效送达。
        </p>
      </DocSection>

      {/* clause-12 */}
      <DocSection id="clause-12" title="第12条 联系我们">
        <p>
          如您对会员服务、本协议条款、退款申请或其他事项有任何疑问，请通过{" "}
          <LegalPlaceholder>{PLACEHOLDERS.supportEmail}</LegalPlaceholder>{" "}
          与我们联系，本平台将在 5 个工作日内予以响应。
        </p>
      </DocSection>

      {/* clause-13 */}
      <DocSection id="clause-13" title="第13条 法律适用与争议解决">
        <p>
          13.1 法律适用：本协议适用中华人民共和国法律（为本协议之目的，不含香港、澳门、台湾）。
        </p>
        <p>
          13.2 争议解决：因本协议产生的争议，双方应首先友好协商解决；协商不成的，任何一方有权向本平台运营方所在地{" "}
          <LegalPlaceholder>{PLACEHOLDERS.courtLocation}</LegalPlaceholder>{" "}
          有管辖权的人民法院提起诉讼。
        </p>
      </DocSection>

      {/* clause-14 */}
      <DocSection id="clause-14" title="第14条 其他">
        <p>
          14.1 本协议以本平台相应页面展示的文本为准，本平台将对协议修改进行存档。
        </p>
        <p>
          14.2 本协议的最终解释权归 <LegalPlaceholder>{PLACEHOLDERS.companyName}</LegalPlaceholder>。
        </p>
      </DocSection>
    </>
  );
}
