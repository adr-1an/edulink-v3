"use client"

import React, {useRef, useState} from "react"
import Link from "next/link"
import { useRouter } from "next/navigation"
import {Camera, ChevronLeft, Eye, Languages, LockKeyhole, Moon, Settings2, Sun, Trash2, UserRound} from "lucide-react"
import { toast } from "sonner"
import LanguageSwitcher from "@/components/app/language-switcher"
import PageTitle, { Subtitle } from "@/components/app/page_title"
import UserAvatar from "@/components/app/user_avatar"

import {
    handleChangePassword,
    handleCompleteProfilePictureUpload,
    handleInitProfilePictureUpload,
    handleRemoveProfilePicture,
    handleSendUpdateEmail,
    handleUpdatePrivacy,
    handleUpdateProfile,
} from "@/app/app/(protected)/staff/profile/actions"
import { handleLogout } from "@/app/app/(protected)/actions"

import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Field, FieldError, FieldLabel } from "@/components/ui/field"
import { Form } from "@/components/ui/form"
import { Input } from "@/components/ui/input"
import {Progress} from "@/components/ui/progress"
import { Switch } from "@/components/ui/switch"
import {Tabs, TabsList, TabsPanel, TabsTab} from "@/components/ui/tabs"
import { useTheme } from "@/components/app/theme_provider"
import {type ProfilePicture} from "@/lib/profile_picture"
import {invalidateProfilePictureCache} from "@/lib/profile_picture_cache"
import {uploadToPresignedURL} from "@/lib/upload_to_presigned_url"
import {useLocale} from "@/i18n/provider"
import TwoFactorSetup, {type TwoFactorStatus} from "./two_factor_setup"
import {
    AlertDialog, AlertDialogClose, AlertDialogDescription, AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogPopup,
    AlertDialogTitle,
    AlertDialogTrigger
} from "@/components/ui/alert-dialog";

export interface StaffProfileUser {
    id: string
    name: string
    profilePicture: ProfilePicture | null
    email: string
    phone: string
    twoFactorStatus: TwoFactorStatus
    publicProfile: boolean
    staffInvitationsDisabled: boolean
    updatedAt: string
    createdAt: string
}

interface Data {
    user: StaffProfileUser
}

interface Props {
    data: Data
}

type Errors = Record<string, string | string[]>

export default function ClientPage({ data }: Props) {
    const router = useRouter()
    const {theme, setTheme} = useTheme()
    const {t} = useLocale()
    const profilePictureInput = useRef<HTMLInputElement>(null)

    const [profileSaving, setProfileSaving] = useState(false)
    const [profilePictureUploading, setProfilePictureUploading] = useState(false)
    const [profilePictureRemoving, setProfilePictureRemoving] = useState(false)
    const [profilePictureProgress, setProfilePictureProgress] = useState(0)
    const [loading2, setLoading2] = useState(false)
    const [loading3, setLoading3] = useState(false)
    const [loading4, setLoading4] = useState(false)
    const [privacySaving, setPrivacySaving] = useState(false)

    const [name, setName] = useState(data.user.name)
    const [phone, setPhone] = useState(data.user.phone)
    const [savedProfile, setSavedProfile] = useState({
        name: data.user.name,
        phone: data.user.phone,
    })
    const [publicProfile, setPublicProfile] = useState(data.user.publicProfile)
    const [staffInvitationsDisabled, setStaffInvitationsDisabled] = useState(
        data.user.staffInvitationsDisabled,
    )

    const [newEmail, setNewEmail] = useState("")
    const [confirmEmail, setConfirmEmail] = useState("")
    const [pass, setPass] = useState("")

    const [currentPass, setCurrentPass] = useState("")
    const [newPass, setNewPass] = useState("")
    const [confirmNewPass, setConfirmNewPass] = useState("")

    const [errors1, setErrors1] = useState<Errors>({})
    const [errors2, setErrors2] = useState<Errors>({})
    const [errors3, setErrors3] = useState<Errors>({})

    async function uploadProfilePicture(event: React.ChangeEvent<HTMLInputElement>) {
        const file = event.target.files?.[0]
        event.target.value = ""
        if (!file || profilePictureUploading || data.user.profilePicture) return

        setProfilePictureUploading(true)
        setProfilePictureProgress(0)

        const initialized = await handleInitProfilePictureUpload({
            name: file.name,
            size: file.size,
            type: file.type,
        })
        if (!initialized.ok) {
            setProfilePictureUploading(false)
            return toast.error(initialized.message)
        }

        try {
            await uploadToPresignedURL(file, initialized.upload.url, setProfilePictureProgress)
        } catch (error) {
            setProfilePictureUploading(false)
            return toast.error(error instanceof Error ? error.message : "The profile picture upload failed.")
        }

        const completed = await handleCompleteProfilePictureUpload(
            initialized.upload.id,
            initialized.upload.completionToken,
        )
        setProfilePictureUploading(false)
        if (!completed.ok) return toast.error(completed.message)

        invalidateProfilePictureCache(`staff:${data.user.id}`)
        toast.success("Profile picture uploaded.")
        router.refresh()
    }

    async function removeProfilePicture() {
        if (profilePictureRemoving) return
        setProfilePictureRemoving(true)
        const result = await handleRemoveProfilePicture()
        setProfilePictureRemoving(false)

        if (!result.ok) return toast.error(result.message)
        invalidateProfilePictureCache(`staff:${data.user.id}`)
        toast.success("Profile picture removed.")
        router.refresh()
    }

    async function onSubmit1(e: React.SubmitEvent<HTMLFormElement>) {
        e.preventDefault()
        setProfileSaving(true)
        setErrors1({})

        const formData = new FormData(e.currentTarget)

        // validate
        let ok = true

        const name = formData.get("name")?.toString() || ""
        if (
            name.length > 64
        ) {
            setErrors1({ root: "The name is too long." })
            ok = false
        }

        const phone = formData.get("phone")?.toString() || ""
        if (
            (phone.length < 3 && phone != "") ||
            phone.length > 64
        ) {
            setErrors1({ root: "Invalid phone number." })
            ok = false
        }

        if (!ok) {
            setProfileSaving(false)
            return
        }

        const res = await handleUpdateProfile(formData, {
            publicProfile,
            staffInvitationsDisabled,
        })

        setProfileSaving(false)

        if (!res.ok) {
            setErrors1({ root: res.message || "Something went wrong." })
            return
        }

        setSavedProfile({ name, phone })
        toast.success("Profile updated.")
    }

    async function onSavePrivacy() {
        setPrivacySaving(true)

        const res = await handleUpdatePrivacy({
            ...savedProfile,
            publicProfile,
            staffInvitationsDisabled,
        })

        setPrivacySaving(false)

        if (!res.ok) {
            toast.error(res.message || "Something went wrong.")
            return
        }

        toast.success("Privacy settings updated.")
    }

    async function onSubmit2(e: React.SubmitEvent<HTMLFormElement>) {
        e.preventDefault()
        setLoading2(true)
        setErrors2({})

        const formData = new FormData(e.currentTarget)

        // validate
        let ok = true

        const email = formData.get("email")?.toString() || ""
        if (
            email.length < 5 ||
            email.length >= 254 ||
            !email.includes("@") ||
            !email.includes(".") ||
            email == data.user.email
        ) {
            setErrors2({ root: "Enter a valid email." })
            ok = false
        }

        const confEmail = formData.get("confirmEmail")?.toString() || ""
        if (
            confEmail !== email
        ) {
            setErrors2({ root: "Emails don't match." })
            ok = false
        }

        if (!ok) {
            setLoading2(false)
            return
        }

        const res = await handleSendUpdateEmail(formData)

        setLoading2(false)

        if (!res.ok) {
            setErrors2({ root: res.message || "Something went wrong." })
            return
        }

        toast.success("Verification link sent. Check the inbox of the email you entered.")
    }

    async function onSubmit3(e: React.SubmitEvent<HTMLFormElement>) {
        e.preventDefault()
        setLoading3(true)
        setErrors3({})

        const formData = new FormData(e.currentTarget)

        // validate
        let ok = true

        const newPass = formData.get("newPassword")?.toString() || ""
        const confPass = formData.get("confirmNewPassword")?.toString() || ""

        if (
            newPass != confPass
        ) {
            setErrors3({ root: "Passwords don't match." })
            ok = false
        } else if (
            newPass.length < 8
        ) {
            setErrors3({ root: "New password is too short." })
            ok = false
        }

        if (!ok) {
            setLoading3(false)
            return
        }

        const res = await handleChangePassword(formData)

        setLoading3(false)

        if (!res.ok) {
            setErrors3({ root: res.message || "Something went wrong." })
            return
        }

        router.refresh()
    }

    async function onLogout() {
        setLoading4(true)

        const res = await handleLogout()

        if (!res.ok) {
            toast.error(res.message || "Something went wrong.")
            setLoading4(false)
            return
        }

        invalidateProfilePictureCache(`staff:${data.user.id}`)
        router.refresh()
    }

    return (
        <div className="mx-auto w-full max-w-4xl">
            <Button variant="ghost" render={<Link href="/app" />}>
                <ChevronLeft /> {t("common.back")}
            </Button>

            <div className="mt-4">
                <PageTitle centered={false}>{t("profile.staff.title")}</PageTitle>
                <p className="mt-1 text-sm text-muted-foreground">{t("profile.staff.description")}</p>
            </div>

            <Tabs className="mt-6 gap-4" defaultValue="profile">
                <div className="overflow-x-auto pb-1">
                    <TabsList className="min-w-max" aria-label={t("profile.staff.sections")}>
                        <TabsTab value="profile"><UserRound /> {t("profile.staff.tabs.profile")}</TabsTab>
                        <TabsTab value="security"><LockKeyhole /> {t("profile.staff.tabs.security")}</TabsTab>
                        <TabsTab value="privacy"><Eye /> {t("profile.staff.tabs.privacy")}</TabsTab>
                        <TabsTab value="preferences"><Settings2 /> {t("profile.staff.tabs.preferences")}</TabsTab>
                    </TabsList>
                </div>

                <TabsPanel className="space-y-4" value="profile">
            <Card className="p-5">
                <div className="flex items-center gap-4">
                    <UserAvatar
                        className="size-20 border shadow-sm"
                        fallbackClassName="text-xl"
                        name={data.user.name}
                        src={data.user.profilePicture?.presignedUrl}
                        cacheKey={`staff:${data.user.id}`}
                    />
                    <div className="min-w-0 flex-1">
                        <p className="font-medium">{t("profile.staff.profilePicture")}</p>
                        <p className="mt-0.5 text-sm text-muted-foreground">
                            {t("profile.staff.profilePictureDescription")}
                        </p>
                        <div className="mt-3 flex flex-wrap gap-2">
                            {!data.user.profilePicture ? (
                                <>
                                    <input
                                        ref={profilePictureInput}
                                        className="sr-only"
                                        type="file"
                                        accept="image/jpeg,image/png,image/gif,image/webp"
                                        disabled={profilePictureUploading}
                                        onChange={uploadProfilePicture}
                                    />
                                    <Button
                                        type="button"
                                        size="sm"
                                        loading={profilePictureUploading}
                                        disabled={profilePictureUploading}
                                        onClick={() => profilePictureInput.current?.click()}
                                    >
                                        <Camera />
                                        {t("profile.staff.choosePicture")}
                                    </Button>
                                </>
                            ) : (
                                <AlertDialog>
                                    <AlertDialogTrigger
                                        render={<Button type="button" size="sm" variant="outline" />}
                                    >
                                        <Trash2 />
                                        {t("profile.staff.remove")}
                                    </AlertDialogTrigger>
                                    <AlertDialogPopup>
                                        <AlertDialogHeader>
                                            <AlertDialogTitle>{t("profile.staff.removePictureTitle")}</AlertDialogTitle>
                                            <AlertDialogDescription>
                                                {t("profile.staff.removePictureDescription")}
                                            </AlertDialogDescription>
                                        </AlertDialogHeader>
                                        <AlertDialogFooter>
                                            <AlertDialogClose render={<Button variant="outline" />}>
                                                {t("profile.staff.cancel")}
                                            </AlertDialogClose>
                                            <Button
                                                type="button"
                                                variant="destructive"
                                                loading={profilePictureRemoving}
                                                disabled={profilePictureRemoving}
                                                onClick={removeProfilePicture}
                                            >
                                                {t("profile.staff.removePicture")}
                                            </Button>
                                        </AlertDialogFooter>
                                    </AlertDialogPopup>
                                </AlertDialog>
                            )}
                        </div>
                        {profilePictureUploading && (
                            <div className="mt-3 space-y-1.5">
                                <Progress value={profilePictureProgress} />
                                <p className="text-xs text-muted-foreground">
                                    {profilePictureProgress < 100
                                        ? `Uploading… ${profilePictureProgress}%`
                                        : "Verifying upload…"}
                                </p>
                            </div>
                        )}
                    </div>
                </div>
            </Card>

            <Card className="p-5 sm:p-8">
                <Subtitle>{t("profile.staff.profileInformation")}</Subtitle>

                <Form errors={errors1} onSubmit={onSubmit1} className="flex flex-col gap-4">
                    {errors1.root && (
                        <p className="text-sm text-destructive">{errors1.root}</p>
                    )}

                    <Field name="email">
                        <FieldLabel>{t("common.email")}</FieldLabel>
                        <Input value={data.user.email} disabled required />
                        <FieldError />
                    </Field>

                    <Field name="name">
                        <FieldLabel>{t("profile.staff.name")}</FieldLabel>
                        <Input
                            name="name"
                            value={name}
                            onChange={(e) => setName(e.target.value)}
                            placeholder="Name"
                            required
                        />
                        <FieldError />
                    </Field>

                    <Field name="phone">
                        <FieldLabel>{t("profile.staff.phoneNumber")}</FieldLabel>
                        <Input
                            name="phone"
                            value={phone}
                            onChange={(e) => setPhone(e.target.value)}
                            placeholder="(555) 123-4567"
                        />
                        <FieldError />
                    </Field>

                    <Button loading={profileSaving} type="submit" disabled={profileSaving}>
                        {t("profile.staff.saveChanges")}
                    </Button>
                </Form>
            </Card>

                </TabsPanel>

                <TabsPanel className="space-y-4" value="preferences">
                    <Card className="p-5">
                        <div className="flex items-center justify-between gap-4">
                            <div>
                                <p className="font-medium">{t("profile.staff.appearance")}</p>
                                <p className="text-sm text-muted-foreground">{t("profile.staff.appearanceDescription")}</p>
                            </div>
                            <div className="flex items-center gap-2 text-muted-foreground">
                                <Sun className="size-4" aria-hidden="true" />
                                <Switch
                                    checked={theme === "dark"}
                                    onCheckedChange={(checked) => setTheme(checked ? "dark" : "light")}
                                    aria-label={t("profile.staff.darkModeAria")}
                                />
                                <Moon className="size-4" aria-hidden="true" />
                            </div>
                        </div>
                    </Card>

                    <Card className="p-5">
                        <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                            <div className="flex items-start gap-3">
                                <span className="flex size-9 shrink-0 items-center justify-center rounded-xl bg-muted text-muted-foreground">
                                    <Languages className="size-4" />
                                </span>
                                <div>
                                    <p className="font-medium">{t("language.label")}</p>
                                    <p className="text-sm text-muted-foreground">{t("language.description")}</p>
                                </div>
                            </div>
                            <LanguageSwitcher className="sm:w-44" />
                        </div>
                    </Card>
                </TabsPanel>

                <TabsPanel className="space-y-4" value="privacy">
            <Card className="p-5 sm:p-8">
                <Subtitle>{t("common.privacy")}</Subtitle>

                <div className="flex flex-col gap-6">
                    <div className="flex items-center justify-between gap-4">
                        <div>
                            <p className="font-medium">{t("profile.staff.publicProfile")}</p>
                            <p className="text-sm text-muted-foreground">
                                {t("profile.staff.publicProfileDescription")}
                            </p>
                        </div>
                        <Switch
                            checked={publicProfile}
                            onCheckedChange={setPublicProfile}
                            disabled={privacySaving}
                            aria-label="Make profile public"
                        />
                    </div>

                    <div className="flex items-center justify-between gap-4">
                        <div>
                            <p className="font-medium">{t("profile.staff.staffInvitations")}</p>
                            <p className="text-sm text-muted-foreground">
                                {t("profile.staff.staffInvitationsDescription")}
                            </p>
                        </div>
                        <Switch
                            checked={!staffInvitationsDisabled}
                            onCheckedChange={(checked) => setStaffInvitationsDisabled(!checked)}
                            disabled={privacySaving}
                            aria-label="Allow staff invitations"
                        />
                    </div>

                    <Button
                        loading={privacySaving}
                        disabled={privacySaving}
                        onClick={onSavePrivacy}
                        type="button"
                    >
                        {t("profile.staff.savePrivacy")}
                    </Button>
                </div>
            </Card>

                </TabsPanel>

                <TabsPanel className="space-y-4" value="security">
            <TwoFactorSetup initialStatus={data.user.twoFactorStatus} />

            <Card className="p-5 sm:p-8">
                <Subtitle>Change email</Subtitle>

                <p className="mb-4 text-center text-muted-foreground">
                    A verification link will be sent to{" "}
                    {confirmEmail === newEmail &&
                    newEmail.length >= 5 &&
                    newEmail.includes("@") &&
                    newEmail !== data.user.email ? (
                        <span className="font-semibold">{newEmail}</span>
                    ) : (
                        "the email you enter"
                    )}
                    . After you click it, the email will be applied automatically.
                </p>

                <Form errors={errors2} onSubmit={onSubmit2} className="flex flex-col gap-4">
                    {errors2.root && (
                        <p className="text-sm text-destructive">{errors2.root}</p>
                    )}

                    <Field name="email">
                        <FieldLabel>New Email</FieldLabel>
                        <Input
                            name="email"
                            value={newEmail}
                            onChange={(e) => setNewEmail(e.target.value)}
                            placeholder="john.doe@example.com"
                            type="email"
                            required
                        />
                        <FieldError />
                    </Field>

                    <Field name="confirmEmail">
                        <FieldLabel>Confirm Email</FieldLabel>
                        <Input
                            value={confirmEmail}
                            onChange={(e) => setConfirmEmail(e.target.value)}
                            placeholder="john.doe@example.com"
                            type="email"
                            required
                        />
                        <FieldError />
                    </Field>

                    <Field name="password">
                        <FieldLabel>Password</FieldLabel>
                        <Input
                            name="password"
                            value={pass}
                            onChange={(e) => setPass(e.target.value)}
                            placeholder="Current Password"
                            type="password"
                            required
                        />
                        <FieldError />
                    </Field>

                    <Button loading={loading2} type="submit" disabled={loading2}>
                        Send verification link
                    </Button>
                </Form>
            </Card>

            <Card className="p-5 sm:p-8">
                <Subtitle>Change password</Subtitle>

                <p className="mb-4 text-center text-muted-foreground">
                    You&apos;ll be logged out of all devices.
                </p>

                <Form errors={errors3} onSubmit={onSubmit3} className="flex flex-col gap-4">
                    {errors3.root && (
                        <p className="text-sm text-destructive">{errors3.root}</p>
                    )}

                    <Field name="password">
                        <FieldLabel>Password</FieldLabel>
                        <Input
                            name="password"
                            value={currentPass}
                            onChange={(e) => setCurrentPass(e.target.value)}
                            placeholder="Current password"
                            type="password"
                            required
                        />
                        <FieldError />
                    </Field>

                    <Field name="newPassword">
                        <FieldLabel>New password</FieldLabel>
                        <Input
                            name="newPassword"
                            value={newPass}
                            onChange={(e) => setNewPass(e.target.value)}
                            placeholder="New password"
                            type="password"
                            required
                        />
                        <FieldError />
                    </Field>

                    <Field name="confirmNewPassword">
                        <FieldLabel>Confirm new password</FieldLabel>
                        <Input
                            value={confirmNewPass}
                            onChange={(e) => setConfirmNewPass(e.target.value)}
                            placeholder="Confirm new password"
                            type="password"
                            required
                        />
                        <FieldError />
                    </Field>

                    <Button type="submit" disabled={loading3} loading={loading3}>
                        Change password
                    </Button>
                </Form>
            </Card>

            <Card className="p-5 sm:p-8">
                <Subtitle>Log out</Subtitle>

                <p className="mb-4 text-center text-muted-foreground">
                    You&apos;ll be logged out on this device only.
                </p>

                <AlertDialog>
                    <AlertDialogTrigger render={<Button />}>
                        Log out
                    </AlertDialogTrigger>

                    <AlertDialogPopup>
                        <AlertDialogHeader>
                            <AlertDialogTitle>Log out</AlertDialogTitle>
                            <AlertDialogDescription>
                                You&#39;ll be logged out on this device only.
                            </AlertDialogDescription>
                        </AlertDialogHeader>

                        <AlertDialogFooter>
                            <AlertDialogClose render={<Button variant="outline" />}>
                                Cancel
                            </AlertDialogClose>

                            <Button
                                loading={loading4}
                                onClick={onLogout}
                                variant="destructive"
                                >
                                Continue
                            </Button>
                        </AlertDialogFooter>
                    </AlertDialogPopup>
                </AlertDialog>
            </Card>
                </TabsPanel>
            </Tabs>
        </div>
    )
}
