import type {Metadata} from "next"
import LegalDocument, {type LegalSection} from "@/components/legal/legal-document"

export const metadata: Metadata = {
    title: "Privacy Policy | Edulink",
    description: "How Vertex collects, uses, shares, and protects information when you use Edulink.",
}

const sections: LegalSection[] = [
    {
        id: "scope-and-roles",
        title: "Scope and our role",
        content: <>
            <p>This Privacy Policy explains how Vertex, based in Warsaw, Poland (&ldquo;Vertex,&rdquo; &ldquo;we,&rdquo; &ldquo;us,&rdquo; or &ldquo;our&rdquo;), processes personal information through Edulink, our websites, applications, and related services (the &ldquo;Service&rdquo;).</p>
            <p>Vertex is the controller of information used to operate Edulink accounts, secure the Service, communicate with users, and manage our relationship with schools. When a school, district, or other educational organization (an &ldquo;Organization&rdquo;) uses Edulink to process student, parent, staff, course, or other education-related information, the Organization generally determines why and how that information is used. For that information, the Organization is generally the controller or responsible educational agency and Vertex acts as its processor or service provider.</p>
            <p>If you use Edulink through an Organization, its own privacy notices and policies may also apply. Questions about an educational record or an Organization&apos;s use of your information should usually be directed to that Organization first.</p>
        </>,
    },
    {
        id: "information-we-collect",
        title: "Information we collect",
        content: <>
            <p>We collect information that you, an Organization, or another authorized user provides, as well as limited technical information generated when the Service is used.</p>
            <h3>Account and profile information</h3>
            <ul>
                <li>Name, email address, optional phone number, password hash, account preferences, and account creation or update dates.</li>
                <li>Privacy choices, including whether a profile is public and whether staff invitations are allowed.</li>
                <li>School memberships, invitations, roles, permissions, and information about who added or invited a user.</li>
            </ul>
            <h3>Organization and educational information</h3>
            <ul>
                <li>School names, regions, academic years, grades, courses, course posts, assignments when available, and related content.</li>
                <li>Student, parent, and staff information supplied by an Organization when those account types and features are available.</li>
                <li>Records of administrative activity, including audit-log entries describing actions taken in a school workspace.</li>
            </ul>
            <h3>Security, device, and usage information</h3>
            <ul>
                <li>Session identifiers, IP address, browser or device user agent, session creation and last-use times, expiry, and revocation information.</li>
                <li>Necessary browser storage and cookies used for authentication, theme preferences, security messages, and remembering the active school workspace within a browser tab.</li>
                <li>Anti-abuse signals processed through Cloudflare Turnstile when you submit protected forms.</li>
            </ul>
            <p>We may receive information from an Organization, an authorized administrator, a person who invites you, or another user who posts content about you. Please do not provide information that is unnecessary for the educational or administrative purpose of the Service.</p>
        </>,
    },
    {
        id: "how-we-use-information",
        title: "How and why we use information",
        content: <>
            <p>We use personal information only as reasonably necessary to:</p>
            <ul>
                <li>Provide accounts, school workspaces, courses, posts, permissions, invitations, and other requested features.</li>
                <li>Authenticate users, maintain sessions, prevent fraud and abuse, investigate security events, and protect users and the Service.</li>
                <li>Send transactional messages such as registration, verification, invitation, password-reset, password-change, and account-security emails.</li>
                <li>Maintain audit trails, troubleshoot problems, provide support, and improve reliability and accessibility.</li>
                <li>Comply with law, enforce our Terms, and establish, exercise, or defend legal claims.</li>
            </ul>
            <p>Where the GDPR or similar law applies, our legal bases may include performance of a contract, steps requested before entering a contract, compliance with legal obligations, legitimate interests in operating and securing the Service, and consent where required. For Organization-controlled educational information, we process the information on the Organization&apos;s documented instructions and the Organization determines the applicable legal basis.</p>
            <p>We do not use student information for targeted advertising, build commercial advertising profiles from it, or sell personal information. We do not use educational information for purposes unrelated to providing, securing, supporting, or improving the school-authorized Service.</p>
        </>,
    },
    {
        id: "children-and-students",
        title: "Children and student information",
        content: <>
            <p>Edulink is intended to be adopted and managed by Organizations. Children should not create an independent account unless the account has been authorized by their Organization, parent, or guardian as required by applicable law.</p>
            <p>When an Organization authorizes Edulink for an educational purpose, Vertex processes student information only to provide that school-authorized service and not for an unrelated commercial purpose. Where applicable, we will rely on school authorization only when the school is legally permitted to act on behalf of a parent or guardian. Otherwise, appropriate parental or guardian consent must be obtained before collecting a child&apos;s personal information.</p>
            <p>Organizations and eligible parents or guardians may ask to review, correct, export, restrict, or delete a child&apos;s information by contacting us. We may need to coordinate the request with the relevant Organization to verify authority and preserve required educational records. We do not knowingly condition a child&apos;s participation on providing more personal information than is reasonably necessary for the educational activity.</p>
        </>,
    },
    {
        id: "sharing",
        title: "How information is shared",
        content: <>
            <p>We may disclose information in the following limited circumstances:</p>
            <ul>
                <li><strong>Within an Organization.</strong> Authorized administrators, staff, students, or parents may access information according to their assigned roles, permissions, and the features used by the Organization.</li>
                <li><strong>Service providers.</strong> Vendors may process information on our behalf for infrastructure, database hosting, email delivery, security, anti-abuse, monitoring, and support. They may use the information only to provide their contracted services and must protect it appropriately.</li>
                <li><strong>Cloudflare Turnstile.</strong> Protected forms use Turnstile to distinguish legitimate users from abusive automated traffic. Cloudflare processes browser and device signals for that purpose under its own Turnstile privacy notice.</li>
                <li><strong>Legal and safety reasons.</strong> We may disclose information when reasonably necessary to comply with law or valid legal process, protect rights and safety, investigate abuse, or prevent fraud and security incidents.</li>
                <li><strong>Business changes.</strong> Information may be transferred as part of a merger, financing, acquisition, reorganization, or sale of assets, subject to appropriate confidentiality and continued protection.</li>
                <li><strong>At your direction.</strong> We may share information when you or the responsible Organization asks us to or provides valid consent.</li>
            </ul>
            <p>We do not sell personal information. We do not share personal information for cross-context behavioral advertising.</p>
        </>,
    },
    {
        id: "visibility-and-controls",
        title: "Visibility and your controls",
        content: <>
            <p>School workspace information is limited by role-based permissions. Organization administrators control staff membership and permissions. Because permissions determine who can view or change school information, administrators should grant only the access required for each person&apos;s role.</p>
            <p>Public profiles are disabled by default. If you enable a public profile, information designated as public may be visible to other people through features that support public profiles. You can turn this setting off from your profile. You can also disable new staff invitations in your privacy settings.</p>
            <p>You are responsible for the information you intentionally include in posts, descriptions, names, and other free-text fields. Avoid including sensitive personal information unless it is necessary, authorized, and appropriate for the educational purpose.</p>
        </>,
    },
    {
        id: "retention",
        title: "Retention and deletion",
        content: <>
            <p>We retain information only for as long as reasonably necessary to provide the Service, meet an Organization&apos;s instructions, maintain security and audit records, comply with law, resolve disputes, and enforce agreements. Retention depends on the type of record, the Organization&apos;s settings and instructions, the sensitivity of the information, and legal requirements.</p>
            <p>Authentication and verification tokens expire or are invalidated when they are used or replaced. Session and security records may remain for a limited period after revocation to detect abuse and investigate incidents. Audit logs may be retained for accountability and school security. Deleted information may remain temporarily in restricted backups until those backups are overwritten under our normal backup cycle.</p>
            <p>When an Organization ends its use of Edulink or asks us to delete Organization-controlled information, we will delete or return that information within a reasonable period unless law or a valid contractual requirement requires retention. Requests can be sent to <a href="mailto:support@vertexapp.net">support@vertexapp.net</a>.</p>
        </>,
    },
    {
        id: "security",
        title: "Security",
        content: <>
            <p>We use reasonable administrative, technical, and organizational safeguards designed to protect information, including hashed passwords, hashed authentication tokens, access controls, role-based permissions, and session revocation. No online service can guarantee absolute security.</p>
            <p>Keep your credentials confidential, use a unique password, sign out of shared devices, and notify us promptly at <a href="mailto:support@vertexapp.net">support@vertexapp.net</a> if you suspect unauthorized access. Organizations should promptly remove access for people who no longer require it and regularly review role permissions.</p>
        </>,
    },
    {
        id: "international-transfers",
        title: "International data transfers",
        content: <>
            <p>Vertex operates from Poland, but our service providers or users may be located in other countries. As a result, information may be processed outside the country where it was collected.</p>
            <p>When required, we use recognized transfer safeguards, such as an adequacy decision, approved contractual clauses, or another lawful transfer mechanism. Organizations may contact us for information relevant to their use of Edulink and international-transfer assessment.</p>
        </>,
    },
    {
        id: "rights",
        title: "Your privacy rights",
        content: <>
            <p>Depending on where you live, you may have rights to access, correct, delete, restrict, or object to processing; receive a portable copy; withdraw consent; and appeal or complain to a privacy regulator. These rights may be limited where an exemption applies.</p>
            <p>For Organization-controlled information, submit the request to your Organization when possible. Vertex will assist the Organization with verified requests. For account information controlled by Vertex, email <a href="mailto:support@vertexapp.net">support@vertexapp.net</a>. We may ask for information needed to verify your identity and authority. Authorized agents must provide proof of authorization.</p>
            <p><strong>EEA, UK, and Switzerland.</strong> You may complain to the supervisory authority where you live or work. In Poland, the supervisory authority is the President of the Personal Data Protection Office (UODO). You may also object to processing based on legitimate interests.</p>
            <p><strong>United States.</strong> Residents of states with comprehensive privacy laws may have applicable access, correction, deletion, portability, opt-out, and appeal rights. Edulink does not sell personal information or use it for targeted advertising. Where FERPA applies and Vertex serves as a school official or contractor, we use education records only for the school-authorized purpose, remain under the school&apos;s control regarding their use and maintenance, and do not redisclose them except as authorized by the school or law.</p>
        </>,
    },
    {
        id: "cookies",
        title: "Cookies and similar technology",
        content: <>
            <p>Edulink uses necessary cookies and browser storage to keep users signed in, protect forms, remember theme preferences, display one-time status messages, and preserve the currently selected school workspace. These technologies are used to provide and secure requested functionality, not for behavioral advertising.</p>
            <p>Blocking necessary cookies or browser storage may prevent authentication, security checks, preferences, or parts of the school navigation from working correctly.</p>
        </>,
    },
    {
        id: "changes-and-contact",
        title: "Changes and contact",
        content: <>
            <p>We may update this Policy as the Service or law changes. We will update the effective date and provide additional notice when a change materially affects how personal information is used or when law requires it. Previous use does not constitute consent where fresh consent is legally required.</p>
            <p>The Service is operated by <strong>Vertex</strong>, Warsaw, Poland. For privacy questions, rights requests, or complaints, contact <a href="mailto:support@vertexapp.net">support@vertexapp.net</a>.</p>
        </>,
    },
]

export default function Page() {
    return (
        <LegalDocument
            title="Privacy Policy"
            summary="This policy explains what information Edulink processes, why it is needed, who can access it, and the choices available to schools, users, parents, and students."
            effectiveDate="July 17 2026"
            sections={sections}
        />
    )
}
