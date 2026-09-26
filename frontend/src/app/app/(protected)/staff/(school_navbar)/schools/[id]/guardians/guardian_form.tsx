"use client"

import {KeyRound} from "lucide-react"
import {Button} from "@/components/ui/button"
import {DialogClose, DialogFooter, DialogPanel} from "@/components/ui/dialog"
import {Field, FieldDescription, FieldLabel} from "@/components/ui/field"
import {Form} from "@/components/ui/form"
import {Input} from "@/components/ui/input"
import {Switch} from "@/components/ui/switch"
import {Textarea} from "@/components/ui/textarea"
import {useLocale} from "@/i18n/provider"
import {isGuardianDraftValid, type GuardianDraft} from "@/lib/guardian"

export default function GuardianForm({draft, saving, mode, wasAccountEnabled = false, onChange, onSubmit}: {
    draft: GuardianDraft
    saving: boolean
    mode: "create" | "edit"
    wasAccountEnabled?: boolean
    onChange: (draft: GuardianDraft) => void
    onSubmit: (event: React.SubmitEvent<HTMLFormElement>) => void
}) {
    const {t} = useLocale()
    const passwordRequired = draft.accountEnabled && (mode === "create" || !wasAccountEnabled)
    const valid = isGuardianDraftValid(draft, mode, wasAccountEnabled)

    return (
        <Form className="contents" onSubmit={onSubmit}>
            <DialogPanel className="grid gap-4 sm:grid-cols-2">
                <Field><FieldLabel>{t("staff.guardians.firstName")}</FieldLabel><Input value={draft.name} onChange={(event) => onChange({...draft, name: event.target.value})} minLength={3} maxLength={32} required autoFocus /></Field>
                <Field><FieldLabel>{t("staff.guardians.lastName")}</FieldLabel><Input value={draft.lastName} onChange={(event) => onChange({...draft, lastName: event.target.value})} minLength={3} maxLength={32} required /></Field>
                <Field><FieldLabel>{t("staff.guardians.dateOfBirth")}</FieldLabel><Input type="date" value={draft.dateOfBirth} onChange={(event) => onChange({...draft, dateOfBirth: event.target.value})} required /></Field>
                <Field><FieldLabel>{t("staff.guardians.email")}</FieldLabel><Input type="email" value={draft.email} onChange={(event) => onChange({...draft, email: event.target.value})} maxLength={254} required /></Field>
                <Field className="sm:col-span-2"><FieldLabel>{t("staff.guardians.phone")}</FieldLabel><Input type="tel" value={draft.phone} onChange={(event) => onChange({...draft, phone: event.target.value})} minLength={draft.phone ? 3 : undefined} maxLength={32} /><FieldDescription>{t("staff.guardians.optional")}</FieldDescription></Field>
                <Field className="sm:col-span-2">
                    <div className="flex items-center justify-between gap-4 rounded-xl border bg-muted/25 p-4">
                        <div><FieldLabel>{t("staff.guardians.guardianLogin")}</FieldLabel><FieldDescription className="mt-1">{t("staff.guardians.guardianLoginDescription")}</FieldDescription></div>
                        <Switch checked={draft.accountEnabled} onCheckedChange={(accountEnabled) => onChange({...draft, accountEnabled, password: accountEnabled ? draft.password : ""})} disabled={saving} aria-label={t("staff.guardians.enableLogin")} />
                    </div>
                </Field>
                <Field className="sm:col-span-2">
                    <FieldLabel>{t(mode === "edit" ? "staff.guardians.newPassword" : "staff.guardians.password")}</FieldLabel>
                    <div className="relative"><KeyRound className="pointer-events-none absolute left-3 top-1/2 z-10 size-4 -translate-y-1/2 text-muted-foreground" /><Input className="pl-9" type="password" value={draft.password} onChange={(event) => onChange({...draft, password: event.target.value})} minLength={8} required={passwordRequired} disabled={!draft.accountEnabled || saving} autoComplete="new-password" /></div>
                    <FieldDescription>{t(passwordRequired ? "staff.guardians.passwordRequired" : mode === "edit" && draft.accountEnabled ? "staff.guardians.passwordKeep" : "staff.guardians.passwordDisabled")}</FieldDescription>
                </Field>
                <Field className="sm:col-span-2">
                    <div className="flex items-center justify-between gap-3"><FieldLabel>{t("staff.guardians.notes")}</FieldLabel><span className="text-xs tabular-nums text-muted-foreground">{draft.notes.length}/2048</span></div>
                    <Textarea className="min-h-28 resize-y" value={draft.notes} onChange={(event) => onChange({...draft, notes: event.target.value})} maxLength={2048} placeholder={t("staff.guardians.notesPlaceholder")} />
                </Field>
            </DialogPanel>
            <DialogFooter>
                <DialogClose render={<Button variant="outline" disabled={saving} />}>{t("staff.guardians.cancel")}</DialogClose>
                <Button type="submit" loading={saving} disabled={!valid || saving}>{t(mode === "create" ? "staff.guardians.add" : "staff.guardians.save")}</Button>
            </DialogFooter>
        </Form>
    )
}
