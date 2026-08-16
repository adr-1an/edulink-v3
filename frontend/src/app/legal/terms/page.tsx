import type {Metadata} from "next"
import LegalDocument, {type LegalSection} from "@/components/legal/legal-document"

export const metadata: Metadata = {
    title: "Terms of Service | Edulink",
    description: "The terms governing access to and use of Edulink.",
}

const sections: LegalSection[] = [
    {
        id: "agreement",
        title: "Agreement and definitions",
        content: <>
            <p>These Terms of Service (&ldquo;Terms&rdquo;) are an agreement between Vertex, based in Warsaw, Poland (&ldquo;Vertex,&rdquo; &ldquo;we,&rdquo; &ldquo;us,&rdquo; or &ldquo;our&rdquo;), and the person or organization that accesses Edulink (&ldquo;you&rdquo;). They govern Edulink&apos;s websites, applications, school workspaces, and related services (the &ldquo;Service&rdquo;).</p>
            <p>An &ldquo;Organization&rdquo; means a school, district, educational institution, or other entity that creates or administers a workspace. An &ldquo;Organization Administrator&rdquo; is a person authorized to manage that workspace. &ldquo;Content&rdquo; means information submitted to the Service, including profile details, school records, grades, courses, posts, assignments, messages, and files when those features are available.</p>
            <p>By creating an account, accepting an invitation, accessing an Organization workspace, or otherwise using the Service, you agree to these Terms. If you use the Service for an Organization, you represent that you are authorized to accept these Terms for that Organization. If you do not agree, do not use the Service.</p>
        </>,
    },
    {
        id: "eligibility",
        title: "Eligibility and child accounts",
        content: <>
            <p>You must be legally capable of entering this agreement. If you are not old enough to enter a binding contract where you live, you may use Edulink only with authorization from a parent, guardian, or Organization that is legally permitted to provide it.</p>
            <p>Accounts for children, including children under 13 in the United States, may be created or used only through a participating Organization or with verifiable parent or guardian authorization where required. A child may not independently represent that they are an Organization administrator or create a school workspace on behalf of a school.</p>
            <p>Vertex remains responsible for its own obligations under applicable children&apos;s privacy law. Nothing in these Terms transfers Vertex&apos;s legal obligations to a school or parent.</p>
        </>,
    },
    {
        id: "accounts",
        title: "Accounts and security",
        content: <>
            <p>You must provide accurate information, keep it reasonably current, and use only an account you are authorized to access. You may not share credentials, sell or transfer an account, impersonate another person, or attempt to bypass invitations, permissions, or authentication controls.</p>
            <p>You are responsible for activity performed through your account until you notify us of suspected compromise. Contact <a href="mailto:support@vertexapp.net">support@vertexapp.net</a> promptly if credentials, invitation links, or security tokens may have been exposed. We may revoke sessions, require a password change, or temporarily restrict an account when reasonably necessary to protect the Service or its users.</p>
            <p>Organization Administrators are responsible for assigning appropriate roles, reviewing permissions, removing access that is no longer needed, and ensuring that administrators are authorized to act for the Organization.</p>
        </>,
    },
    {
        id: "organization-responsibilities",
        title: "Organization responsibilities",
        content: <>
            <p>An Organization controls the educational and administrative purposes for which it uses Edulink. The Organization must:</p>
            <ul>
                <li>Have lawful authority to submit and direct the processing of Content, including student and parent information.</li>
                <li>Provide required notices and obtain any authorization or consent required for its use of the Service.</li>
                <li>Use student information only for legitimate educational and administrative purposes.</li>
                <li>Configure roles and permissions appropriately and respond to requests concerning educational records.</li>
                <li>Ensure its instructions to Vertex comply with applicable education, employment, privacy, accessibility, and records-retention laws.</li>
            </ul>
            <p>Where an Organization is permitted to authorize collection from a child on behalf of a parent, Edulink may rely on that authorization only for the Organization-approved educational purpose. The Organization may request access to, correction of, export of, or deletion of the student information it controls, subject to applicable law.</p>
        </>,
    },
    {
        id: "acceptable-use",
        title: "Acceptable use",
        content: <>
            <p>You may use the Service only for lawful educational, administrative, and related purposes. You must not:</p>
            <ul>
                <li>Access or disclose information without authorization, exceed assigned permissions, or attempt to identify security vulnerabilities without our written permission.</li>
                <li>Upload malware, disrupt the Service, automate abusive traffic, evade rate limits, or interfere with another user&apos;s access.</li>
                <li>Harass, threaten, exploit, discriminate against, or endanger another person, particularly a child.</li>
                <li>Post unlawful, fraudulent, defamatory, infringing, sexually exploitative, or otherwise harmful Content.</li>
                <li>Use Edulink for surveillance, targeted advertising, commercial profiling of students, or decisions prohibited by law.</li>
                <li>Scrape, sell, rent, sublicense, reverse engineer, or commercially exploit the Service except where applicable law does not allow that restriction.</li>
                <li>Submit highly sensitive information that the applicable feature and Organization have not approved, including unnecessary medical, biometric, financial, or government-identifier data.</li>
            </ul>
            <p>We may investigate suspected violations and remove Content or restrict access when reasonably necessary. When appropriate, we will work with the relevant Organization before acting on Organization-controlled Content.</p>
        </>,
    },
    {
        id: "content",
        title: "Content and ownership",
        content: <>
            <p>You or the relevant Organization retain ownership of Content submitted to Edulink. You grant Vertex a limited, worldwide, non-exclusive license to host, copy, transmit, display, format, back up, and otherwise process Content only as necessary to provide, secure, support, and improve the Service, comply with law, and follow the Organization&apos;s documented instructions.</p>
            <p>You represent that you have the rights and authority needed to submit Content and permit this processing. You are responsible for the accuracy, legality, and appropriateness of Content you submit.</p>
            <p>Vertex and its licensors retain all rights in the Service, including its software, interface, design, documentation, trademarks, and non-user-created materials. These Terms give you a limited, revocable, non-transferable right to use the Service; they do not transfer ownership of the Service or Vertex branding.</p>
            <p>If you voluntarily send feedback, you permit us to use it without restriction or compensation, provided we do not publicly identify you without permission.</p>
        </>,
    },
    {
        id: "education-data",
        title: "Education data commitments",
        content: <>
            <p>Vertex will process Organization-controlled education data only to provide, secure, maintain, and support the school-authorized Service; follow lawful Organization instructions; and meet legal obligations.</p>
            <p>We will not sell student information, use it for targeted advertising, or create commercial profiles unrelated to the educational service. We will not disclose education records except to authorized Organization users, contracted service providers supporting Edulink, or as directed by the Organization or required by law.</p>
            <p>Where FERPA applies and the parties rely on the school-official exception, Vertex will perform an institutional service for which the Organization would otherwise use its employees, remain under the Organization&apos;s direct control regarding education records, use those records only for the authorized purpose, and avoid unauthorized redisclosure. These Terms do not by themselves determine whether a particular Organization&apos;s use satisfies FERPA or another local education law.</p>
        </>,
    },
    {
        id: "privacy",
        title: "Privacy",
        content: <>
            <p>Our <a href="/legal/privacy">Privacy Policy</a> explains how we collect, use, disclose, retain, and protect personal information. It is incorporated into these Terms by reference.</p>
            <p>Each party will comply with the privacy and security obligations that apply to its role. If an Organization requires a data processing agreement, student privacy agreement, security schedule, or international-transfer terms, those must be agreed separately in writing. If a signed agreement conflicts with these online Terms, the signed agreement controls for that conflict.</p>
        </>,
    },
    {
        id: "third-party-services",
        title: "Third-party services",
        content: <>
            <p>Edulink depends on third-party infrastructure and service providers, including providers for hosting, email delivery, databases, and bot protection. Their availability may affect the Service. Our use of providers does not reduce our responsibility to select and manage them appropriately.</p>
            <p>The Service may link to third-party websites. Vertex does not control those websites and is not responsible for their content or practices. A link does not imply endorsement.</p>
        </>,
    },
    {
        id: "service-changes",
        title: "Availability and changes",
        content: <>
            <p>We aim to keep Edulink reliable, but the Service may occasionally be unavailable because of maintenance, security incidents, provider failures, testing, or events outside our reasonable control. Unless a separate written agreement states otherwise, we do not promise a particular uptime or support-response time.</p>
            <p>We may add, modify, or discontinue features. We will try to provide reasonable notice before a material discontinuation when practical. Beta, preview, or experimental features may be incomplete, change without notice, and be subject to additional terms.</p>
            <p>Edulink is an organizational tool. It is not an emergency service, a substitute for professional educational, medical, legal, or safeguarding judgment, or the authoritative source for decisions where law requires human review. Organizations should maintain appropriate backup and continuity procedures for critical records.</p>
        </>,
    },
    {
        id: "suspension-and-termination",
        title: "Suspension and termination",
        content: <>
            <p>You may stop using Edulink at any time. An Organization may remove a user from its workspace. To request closure of an account or Organization workspace, contact <a href="mailto:support@vertexapp.net">support@vertexapp.net</a>.</p>
            <p>We may suspend or terminate access when reasonably necessary because of a material breach, security risk, unlawful use, non-payment under a future paid plan, legal requirement, or risk of harm to users or the Service. Where appropriate, we will provide notice and an opportunity to remedy the issue.</p>
            <p>After termination, access ends. We will handle Content according to the Privacy Policy, the Organization&apos;s instructions, applicable law, and any separate written agreement. Provisions that by their nature should survive—including ownership, confidentiality, disclaimers, liability limits, and dispute terms—will survive.</p>
        </>,
    },
    {
        id: "disclaimers",
        title: "Disclaimers",
        content: <>
            <p>To the fullest extent permitted by law, the Service is provided &ldquo;as is&rdquo; and &ldquo;as available.&rdquo; Vertex disclaims implied warranties of merchantability, fitness for a particular purpose, non-infringement, and uninterrupted or error-free operation.</p>
            <p>Nothing in these Terms excludes warranties, guarantees, or other rights that cannot lawfully be excluded, including mandatory consumer protections. Vertex does not warrant that user-provided Content is accurate or that a particular Organization has configured permissions correctly.</p>
        </>,
    },
    {
        id: "liability",
        title: "Limitation of liability",
        content: <>
            <p>Nothing in these Terms limits liability that cannot legally be limited, including liability for intentional misconduct, gross negligence where applicable, fraud, or death or personal injury caused by negligence.</p>
            <p>To the fullest extent permitted by law, Vertex will not be liable for indirect, incidental, special, consequential, or punitive damages, or for lost profits, revenue, goodwill, or data, arising from the Service. Vertex&apos;s total liability arising from the Service during any twelve-month period will not exceed the greater of the amount paid to Vertex for the Service during that period or EUR 100.</p>
            <p>These limitations apply only to the extent permitted by applicable law. They do not deprive consumers of mandatory protections available in their country of habitual residence.</p>
        </>,
    },
    {
        id: "indemnity",
        title: "Organization indemnity",
        content: <>
            <p>To the extent permitted by law, an Organization will defend and indemnify Vertex against third-party claims resulting from the Organization&apos;s unlawful Content, instructions, use of the Service, or material breach of these Terms. This obligation does not apply to the extent a claim was caused by Vertex&apos;s breach, negligence, or unlawful conduct.</p>
            <p>This section does not apply to individual consumers acting outside a trade, business, craft, or profession.</p>
        </>,
    },
    {
        id: "law-and-disputes",
        title: "Governing law and disputes",
        content: <>
            <p>Before filing a claim, please contact <a href="mailto:support@vertexapp.net">support@vertexapp.net</a> so we can try to resolve the issue informally.</p>
            <p>These Terms are governed by Polish law, without regard to conflict-of-law principles. Courts with jurisdiction in Warsaw, Poland will have jurisdiction over disputes, except where mandatory law gives you the right to bring a claim elsewhere.</p>
            <p>If you are a consumer, this choice does not deprive you of mandatory protections provided by the law of your country of habitual residence, and any mandatory consumer forum rights remain available. Nothing in these Terms restricts a right to complain to a regulator or supervisory authority.</p>
        </>,
    },
    {
        id: "general",
        title: "General terms",
        content: <>
            <p>We may update these Terms to reflect legal, security, or Service changes. We will update the effective date and provide reasonable notice of material changes. If a change requires consent, we will request it. Continued use after non-consent-based changes take effect constitutes acceptance where permitted by law.</p>
            <p>You may not assign these Terms without our consent. We may assign them as part of a merger, reorganization, financing, or sale of the relevant business, provided your rights are not materially reduced. Our failure to enforce a provision is not a waiver. If a provision is unenforceable, it will be limited to the minimum extent necessary and the remainder will continue.</p>
            <p>These Terms, the Privacy Policy, and any applicable signed agreement form the agreement concerning the Service. Headings are for convenience only.</p>
        </>,
    },
    {
        id: "contact",
        title: "Contact",
        content: <>
            <p>Edulink is operated by <strong>Vertex</strong>, Warsaw, Poland.</p>
            <p>Questions about these Terms may be sent to <a href="mailto:support@vertexapp.net">support@vertexapp.net</a>.</p>
        </>,
    },
]

export default function Page() {
    return (
        <LegalDocument
            title="Terms of Service"
            summary="These terms set the rules for using Edulink and explain the responsibilities of Vertex, schools, administrators, staff, students, parents, and other users."
            effectiveDate="July 17 2026"
            sections={sections}
        />
    )
}
