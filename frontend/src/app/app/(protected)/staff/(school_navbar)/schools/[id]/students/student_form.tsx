"use client"

import {KeyRound, Send} from "lucide-react"
import {Button} from "@/components/ui/button"
import {DialogClose, DialogFooter, DialogPanel} from "@/components/ui/dialog"
import {Field, FieldDescription, FieldLabel} from "@/components/ui/field"
import {Form} from "@/components/ui/form"
import {Input} from "@/components/ui/input"
import {Switch} from "@/components/ui/switch"
import {Textarea} from "@/components/ui/textarea"
import {useLocale} from "@/i18n/provider"
import {
    emptyStudentDraft, studentDraftToInput, type StudentDraft,
} from "@/lib/student"
import {canSendPortalUserActivationLink} from "@/lib/portal_user_activation_link"

export {emptyStudentDraft, studentDraftToInput}
export type {StudentDraft}

export default function StudentForm({draft, saving, sendingActivationLink = false, activationLinkSent = false, canSendActivationLink = false, mode, wasAccountEnabled = false, onChange, onSubmit, onSendActivationLink}: {
    draft: StudentDraft
    saving: boolean
    sendingActivationLink?: boolean
    activationLinkSent?: boolean
    canSendActivationLink?: boolean
    mode: "create" | "edit"
    wasAccountEnabled?: boolean
    onChange: (draft: StudentDraft) => void
    onSubmit: (event: React.SubmitEvent<HTMLFormElement>) => void
    onSendActivationLink?: () => void
}) {
    const {t} = useLocale()
    const passwordRequired = draft.accountEnabled && (mode === "create" || !wasAccountEnabled)
    const valid = draft.name.trim().length >= 3 && draft.name.trim().length <= 32
        && draft.lastName.trim().length >= 3 && draft.lastName.trim().length <= 32
        && draft.email.trim().length >= 5 && draft.email.trim().length <= 254
        && (!draft.phone.trim() || (draft.phone.trim().length >= 3 && draft.phone.trim().length <= 32))
        && draft.notes.trim().length <= 2048
        && (!draft.password || draft.password.length >= 8)
        && (!passwordRequired || draft.password.length >= 8)
    const activationLinkEnabled = canSendPortalUserActivationLink({
        accountEnabled: draft.accountEnabled,
        accountActive: draft.accountActive,
        password: draft.password,
        activationLinkSent,
    })

    return (
        <Form className="contents" onSubmit={onSubmit}>
            <DialogPanel className="grid gap-4 sm:grid-cols-2">
                <Field><FieldLabel>{t("staff.students.firstName")}</FieldLabel><Input value={draft.name} onChange={(event) => onChange({...draft, name: event.target.value})} minLength={3} maxLength={32} required autoFocus /></Field>
                <Field><FieldLabel>{t("staff.students.lastName")}</FieldLabel><Input value={draft.lastName} onChange={(event) => onChange({...draft, lastName: event.target.value})} minLength={3} maxLength={32} required /></Field>
                <Field><FieldLabel>{t("staff.students.dateOfBirth")}</FieldLabel><Input type="date" value={draft.dateOfBirth} onChange={(event) => onChange({...draft, dateOfBirth: event.target.value})} required /></Field>
                <Field><FieldLabel>{t("staff.students.email")}</FieldLabel><Input type="email" value={draft.email} onChange={(event) => onChange({...draft, email: event.target.value})} maxLength={254} required /></Field>
                <Field className="sm:col-span-2"><FieldLabel>{t("staff.students.phone")}</FieldLabel><Input type="tel" value={draft.phone} onChange={(event) => onChange({...draft, phone: event.target.value})} minLength={draft.phone ? 3 : undefined} maxLength={32} /><FieldDescription>{t("staff.students.optional")}</FieldDescription></Field>
                <Field className="sm:col-span-2">
                    <div className="rounded-xl border bg-muted/25">
                        <div className="flex items-center justify-between gap-4 p-4">
                            <div><FieldLabel>{t("staff.students.studentLogin")}</FieldLabel><FieldDescription className="mt-1">{t("staff.students.studentLoginDescription")}</FieldDescription></div>
                            <Switch checked={draft.accountEnabled} onCheckedChange={(accountEnabled) => onChange({...draft, accountEnabled, accountActive: accountEnabled ? draft.accountActive : false, password: accountEnabled ? draft.password : ""})} disabled={saving} aria-label={t("staff.students.enableLogin")} />
                        </div>
                        {draft.accountEnabled && (
                            <div className="flex items-center justify-between gap-4 border-t p-4">
                                <div><FieldLabel>{t("staff.students.accountActive")}</FieldLabel><FieldDescription className="mt-1">{t("staff.students.accountActiveDescription")}</FieldDescription></div>
                                <Switch checked={draft.accountActive} onCheckedChange={(accountActive) => onChange({...draft, accountActive})} disabled={saving} aria-label={t("staff.students.accountActive")} />
                            </div>
                        )}
                    </div>
                </Field>
                <Field className="sm:col-span-2">
                    <FieldLabel>{t(mode === "edit" ? "staff.students.newPassword" : "staff.students.password")}</FieldLabel>
                    <div className="relative"><KeyRound className="pointer-events-none absolute left-3 top-1/2 z-10 size-4 -translate-y-1/2 text-muted-foreground" /><Input className="pl-9" type="password" value={draft.password} onChange={(event) => onChange({...draft, password: event.target.value})} minLength={8} required={passwordRequired} disabled={!draft.accountEnabled || saving} autoComplete="new-password" /></div>
                    <FieldDescription>{t(passwordRequired ? "staff.students.passwordRequired" : mode === "edit" && draft.accountEnabled ? "staff.students.passwordKeep" : "staff.students.passwordDisabled")}</FieldDescription>
                    {mode === "edit" && canSendActivationLink && onSendActivationLink && (
                        <div className="mt-2 flex flex-wrap items-center justify-between gap-3 rounded-lg border bg-muted/20 p-3">
                            <p className="text-xs text-muted-foreground">{t("staff.portalUsers.activation.description")}</p>
                            <Button type="button" size="sm" variant="outline" loading={sendingActivationLink} disabled={!activationLinkEnabled || saving || sendingActivationLink} onClick={onSendActivationLink}>
                                <Send /> {t("staff.portalUsers.activation.send")}
                            </Button>
                        </div>
                    )}
                </Field>
                <Field className="sm:col-span-2">
                    <div className="flex items-center justify-between gap-3"><FieldLabel>{t("staff.students.notes")}</FieldLabel><span className="text-xs tabular-nums text-muted-foreground">{draft.notes.length}/2048</span></div>
                    <Textarea className="min-h-28 resize-y" value={draft.notes} onChange={(event) => onChange({...draft, notes: event.target.value})} maxLength={2048} placeholder={t("staff.students.notesPlaceholder")} />
                </Field>
            </DialogPanel>
            <DialogFooter>
                <DialogClose render={<Button variant="outline" disabled={saving} />}>{t("staff.students.cancel")}</DialogClose>
                <Button type="submit" loading={saving} disabled={!valid || saving}>{t(mode === "create" ? "staff.students.add" : "staff.students.save")}</Button>
            </DialogFooter>
        </Form>
    )
}
