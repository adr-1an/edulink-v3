import {cookies} from "next/headers"
import {redirect} from "next/navigation"
import {
    AtSign, Building2, CalendarDays, CircleX, Globe2, GlobeX, Info,
    Languages, Mail, Phone, School, ServerCrash, ShieldCheck, UserRound,
} from "lucide-react"
import ErrorPage from "@/components/app/error"
import LanguageSwitcher from "@/components/app/language-switcher"
import UserAvatar from "@/components/app/user_avatar"
import {Badge} from "@/components/ui/badge"
import {Button} from "@/components/ui/button"
import {Card, CardContent, CardHeader, CardTitle} from "@/components/ui/card"
import {getTranslations} from "@/i18n/server"
import {parsePortalProfile} from "@/lib/portal_profile"
import PasswordUpdateCard from "./password_update_card"

type AccountType = "student" | "guardian"

function formatDateOfBirth(value: string, locale: string, fallback: string) {
    const datePart = value.match(/^\d{4}-\d{2}-\d{2}/)?.[0]
    if (!datePart) return fallback
    const date = new Date(`${datePart}T00:00:00Z`)
    if (Number.isNaN(date.getTime())) return fallback
    return new Intl.DateTimeFormat(locale === "pl" ? "pl-PL" : "en-US", {
        day: "numeric",
        month: "long",
        year: "numeric",
        timeZone: "UTC",
    }).format(date)
}

export default async function PortalProfilePage({accountType}: {accountType: AccountType}) {
    const {locale, t} = await getTranslations()
    const copy = accountType === "guardian"
        ? {
            title: t("profile.guardian.title"),
            managedTitle: t("profile.guardian.managedTitle"),
            managedDescription: t("profile.guardian.managedDescription"),
            personalInformation: t("profile.guardian.personalInformation"),
            fullName: t("profile.guardian.fullName"),
            dateOfBirth: t("profile.guardian.dateOfBirth"),
            yourSchool: t("profile.guardian.yourSchool"),
            schoolAdministrator: t("profile.guardian.schoolAdministrator"),
            contactAdministrator: t("profile.guardian.contactAdministrator"),
            networkError: t("profile.guardian.error.network"),
            serverError: t("profile.guardian.error.server"),
            loadError: t("profile.guardian.error.load"),
            invalidError: t("profile.guardian.error.invalid"),
            incompleteError: t("profile.guardian.error.incomplete"),
        }
        : {
            title: t("profile.student.title"),
            managedTitle: t("profile.student.managedTitle"),
            managedDescription: t("profile.student.managedDescription"),
            personalInformation: t("profile.student.personalInformation"),
            fullName: t("profile.student.fullName"),
            dateOfBirth: t("profile.student.dateOfBirth"),
            yourSchool: t("profile.student.yourSchool"),
            schoolAdministrator: t("profile.student.schoolAdministrator"),
            contactAdministrator: t("profile.student.contactAdministrator"),
            networkError: t("profile.student.error.network"),
            serverError: t("profile.student.error.server"),
            loadError: t("profile.student.error.load"),
            invalidError: t("profile.student.error.invalid"),
            incompleteError: t("profile.student.error.incomplete"),
        }
    const token = (await cookies()).get("portal_token")?.value
    let res: Response

    try {
        res = await fetch(`${process.env.API_URL}/v1/portal/profile`, {
            headers: {Authorization: `Bearer ${token}`},
            cache: "no-store",
        })
    } catch {
        return <ErrorPage message={copy.networkError} icon={GlobeX} />
    }

    if (res.status === 401) redirect("/app/portal")
    if (!res.ok) {
        return <ErrorPage message={res.status === 500 ? copy.serverError : copy.loadError} icon={res.status === 500 ? ServerCrash : CircleX} />
    }

    let data: unknown
    try {
        data = await res.json()
    } catch {
        return <ErrorPage message={copy.invalidError} icon={CircleX} />
    }

    if (!data || typeof data !== "object" || !("profile" in data)) {
        return <ErrorPage message={copy.incompleteError} icon={CircleX} />
    }
    const profile = parsePortalProfile(data.profile)
    if (!profile) return <ErrorPage message={copy.incompleteError} icon={CircleX} />

    const fullName = `${profile.name} ${profile.lastName}`.trim()
    return (
        <div className="space-y-6">
            <header className="flex flex-col justify-between gap-5 rounded-2xl border bg-card p-5 sm:flex-row sm:items-center sm:p-7">
                <div className="flex min-w-0 items-center gap-4">
                    <UserAvatar
                        className="size-14 border sm:size-16"
                        fallbackClassName="text-lg font-semibold"
                        name={fullName}
                        src={profile.profilePictureURL}
                        cacheKey="portal:current-user"
                    />
                    <div className="min-w-0">
                        <div className="flex flex-wrap items-center gap-2">
                            <h1 className="truncate text-2xl font-semibold tracking-tight sm:text-3xl">{fullName}</h1>
                            <Badge variant="secondary"><ShieldCheck /> {t("common.readOnly")}</Badge>
                        </div>
                        <p className="mt-1 text-sm text-muted-foreground">{copy.title}</p>
                    </div>
                </div>
                <Badge className="w-fit" variant="outline">
                    <UserRound /> {t(accountType === "guardian" ? "common.guardian" : "common.student")}
                </Badge>
            </header>

            <div className="flex gap-3 rounded-2xl border border-info/20 bg-info/8 p-4 text-sm">
                <Info className="mt-0.5 size-5 shrink-0 text-info-foreground" />
                <div>
                    <p className="font-medium">{copy.managedTitle}</p>
                    <p className="mt-1 text-muted-foreground">{copy.managedDescription}</p>
                </div>
            </div>

            <Card>
                <CardContent className="flex flex-col gap-4 py-5 sm:flex-row sm:items-center sm:justify-between">
                    <div className="flex items-start gap-3">
                        <span className="flex size-9 shrink-0 items-center justify-center rounded-xl bg-muted text-muted-foreground"><Languages className="size-4" /></span>
                        <div>
                            <p className="font-medium">{t("language.label")}</p>
                            <p className="text-sm text-muted-foreground">{t("language.description")}</p>
                        </div>
                    </div>
                    <LanguageSwitcher className="sm:w-48" />
                </CardContent>
            </Card>

            <PasswordUpdateCard />

            <div className="grid gap-5 lg:grid-cols-[minmax(0,1.4fr)_minmax(19rem,0.8fr)]">
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2 text-lg"><UserRound className="size-5" /> {copy.personalInformation}</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <dl className="grid gap-x-6 sm:grid-cols-2">
                            <ProfileField icon={UserRound} label={copy.fullName} value={fullName} />
                            <ProfileField icon={Mail} label={t("common.email")} value={profile.email} />
                            <ProfileField icon={Phone} label={t("common.phone")} value={profile.phone || t("common.notProvided")} />
                            <ProfileField icon={CalendarDays} label={copy.dateOfBirth} value={formatDateOfBirth(profile.dateOfBirth, locale, t("common.notProvided"))} />
                        </dl>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2 text-lg"><School className="size-5" /> {copy.yourSchool}</CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-5">
                        <div className="space-y-3">
                            <InfoRow icon={Building2} label={t("common.school")} value={profile.school.name} />
                            <InfoRow icon={Globe2} label={t("common.region")} value={profile.school.region || t("common.notProvided")} />
                        </div>
                        <div className="rounded-2xl border bg-muted/30 p-4">
                            <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">{copy.schoolAdministrator}</p>
                            <p className="mt-2 font-medium">{profile.school.owner.name}</p>
                            <p className="mt-0.5 break-all text-sm text-muted-foreground">{profile.school.owner.email}</p>
                            <Button className="mt-4 w-full" variant="outline" render={<a href={`mailto:${profile.school.owner.email}`} />}><AtSign /> {copy.contactAdministrator}</Button>
                        </div>
                    </CardContent>
                </Card>
            </div>
        </div>
    )
}

function ProfileField({icon: Icon, label, value}: {icon: typeof UserRound; label: string; value: string}) {
    return (
        <div className="border-b py-4 last:border-b-0 sm:[&:nth-last-child(-n+2)]:border-b-0">
            <dt className="flex items-center gap-2 text-xs font-medium text-muted-foreground"><Icon className="size-3.5" /> {label}</dt>
            <dd className="mt-1.5 break-words text-sm font-medium">{value}</dd>
        </div>
    )
}

function InfoRow({icon: Icon, label, value}: {icon: typeof Building2; label: string; value: string}) {
    return (
        <div className="flex items-start gap-3">
            <span className="flex size-9 shrink-0 items-center justify-center rounded-xl bg-muted text-muted-foreground"><Icon className="size-4" /></span>
            <div className="min-w-0">
                <p className="text-xs text-muted-foreground">{label}</p>
                <p className="mt-0.5 break-words text-sm font-medium">{value}</p>
            </div>
        </div>
    )
}
