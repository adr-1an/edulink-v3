"use client"

import React, {useRef, useState} from "react"
import {Reorder, useDragControls} from "framer-motion"
import {useRouter} from "next/navigation"
import {
    BookOpen, CircleHelp, Crown, Eye, GripVertical, KeyRound, Pencil, Plus, ShieldCheck,
    Trash2, TriangleAlert, UserCog, type LucideIcon,
} from "lucide-react"
import {toast} from "sonner"
import PageTitle from "@/components/app/page_title"
import LocalDateTime from "@/components/local-date-time"
import {
    AlertDialog, AlertDialogClose, AlertDialogDescription, AlertDialogFooter,
    AlertDialogHeader, AlertDialogPopup, AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import {Badge} from "@/components/ui/badge"
import {Button} from "@/components/ui/button"
import {Card} from "@/components/ui/card"
import {
    ContextMenu, ContextMenuItem, ContextMenuPopup, ContextMenuTrigger,
} from "@/components/ui/context-menu"
import {
    Dialog, DialogClose, DialogDescription, DialogFooter, DialogHeader,
    DialogPanel, DialogPopup, DialogTitle, DialogTrigger,
} from "@/components/ui/dialog"
import {Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle} from "@/components/ui/empty"
import {Field, FieldDescription, FieldLabel} from "@/components/ui/field"
import {Form} from "@/components/ui/form"
import {Input} from "@/components/ui/input"
import {
    NumberField, NumberFieldDecrement, NumberFieldGroup,
    NumberFieldIncrement, NumberFieldInput,
} from "@/components/ui/number-field"
import {Switch} from "@/components/ui/switch"
import {canManagePosition, emptySchoolAccess, hasSchoolPermission, highestRolePosition, type SchoolAccess} from "@/lib/school_access"
import {
    handleCreateRole, handleDeleteRole, handleReorderRoles,
    handleSetRolePermission, handleUpdateRole,
} from "./actions"
import {useLocale} from "@/i18n/provider"
import {
    localizedPermissionCategory, localizedPermissionDescription,
    localizedPermissionLabel, localizedPermissionPreset,
} from "@/i18n/role-permissions"

interface Role {
    id: string
    position: number
    name: string
    color: string
    createdAt: string
    permissions: string[]
}

const permissionLabels: Record<string, string> = {
    "school.view": "View school",
    "school.update": "Edit school settings",
    "school.promote": "Promote students",
    "school.invite.list": "View staff invitations",
    "school.invite.cancel": "Cancel staff invitations",
    "staff.view": "View staff",
    "staff.create": "Invite staff",
    "staff.delete": "Remove staff",
    "staff.role.add": "Assign staff roles",
    "staff.role.remove": "Remove staff roles",
    "staff.role.list": "View staff roles",
    "academicYear.create": "Create academic years",
    "academicYear.list": "View academic years",
    "academicYear.toggleActive": "Change active academic year",
    "academicYear.delete": "Delete academic years",
    "grade.list": "View grades",
    "grade.create": "Create grades",
    "grade.update": "Edit grades",
    "grade.delete": "Delete grades",
    "role.create": "Create roles",
    "role.list": "View roles",
    "role.update": "Edit roles",
    "role.delete": "Delete roles",
    "role.permission.update": "Manage role permissions",
    "course.create": "Create course",
    "course.list": "View courses",
    "course.update": "Edit course",
    "course.delete": "Delete course",
    "course.post.create": "Create course posts",
    "course.post.list": "View course posts",
    "course.post.view": "Open course posts",
    "course.post.update": "Edit course posts",
    "course.post.delete": "Delete course posts",
    "post.attachment.create": "Upload post attachments",
    "post.attachment.delete": "Delete post attachments",
    "course.assignment.create": "Create course assignments",
    "course.assignment.list": "View course assignments",
    "course.assignment.update": "Edit course assignments",
    "course.assignment.delete": "Delete course assignments",
    "submission.list": "View assignment submissions",
    "submission.view": "Open assignment submissions",
    "submission.return": "Return assignment submissions",
    "submission.delete": "Delete returned submissions",
    "submission.grade": "Grade assignment submissions",
    "submission.removeGrade": "Remove assignment grades",
    "course.student.list": "View course students",
    "course.student.assign": "Assign course students",
    "course.student.remove": "Remove course students",
    "student.list": "View students",
    "student.view": "View student profiles",
    "student.create": "Create students",
    "student.update": "Edit students",
    "student.delete": "Delete students",
    "log.list": "View audit logs",
}

const permissionDescriptions: Record<string, string> = {
    "school.view": "Can open the school dashboard and view school information.",
    "school.update": "Can change the school's name, region, and other settings.",
    "school.delete": "Can delete the school and remove access to it.",
    "school.promote": "Can promote the school into its next academic year.",
    "school.invite.list": "Can view pending and rejected staff invitations for the school.",
    "school.invite.cancel": "Can cancel pending staff invitations before they are accepted.",
    "staff.view": "Can view the school's staff members and their assigned roles.",
    "staff.create": "Can invite new staff members to join the school.",
    "staff.delete": "Can remove staff whose highest role is below their own highest role, but cannot remove themselves.",
    "staff.role.add": "Can assign roles below their own highest role to staff members.",
    "staff.role.remove": "Can remove roles below their own highest role from staff members.",
    "staff.role.list": "Can view the roles assigned to staff members.",
    "academicYear.create": "Can create new academic years for the school.",
    "academicYear.list": "Can view the school's academic years.",
    "academicYear.toggleActive": "Can activate an academic year or clear the currently active one.",
    "academicYear.delete": "Can delete academic years that are not currently active.",
    "grade.list": "Can view grades in the school's academic years.",
    "grade.create": "Can create new grades inside an academic year.",
    "grade.update": "Can change a grade's name and level.",
    "grade.delete": "Can permanently delete grades.",
    "role.create": "Can create roles below their own highest role.",
    "role.list": "Can view the school's roles and their hierarchy.",
    "role.update": "Can edit and reorder roles below their own highest role.",
    "role.delete": "Can delete roles below their own highest role.",
    "role.permission.update": "Can change permissions only for roles below their own highest role.",
    "course.create": "Can create new courses inside a grade.",
    "course.list": "Can view courses inside a grade.",
    "course.update": "Can edit a course's name, description and color.",
    "course.delete": "Can permanently delete courses.",
    "course.post.create": "Can publish posts inside courses.",
    "course.post.list": "Can view posts inside courses.",
    "course.post.view": "Can open individual course posts in the full post reader.",
    "course.post.update": "Can edit existing course posts.",
    "course.post.delete": "Can permanently delete course posts.",
    "post.attachment.create": "Can upload and attach files to course posts.",
    "post.attachment.delete": "Can permanently remove files attached to course posts.",
    "course.assignment.create": "Can create assignments inside courses.",
    "course.assignment.list": "Can view assignments inside courses.",
    "course.assignment.update": "Can edit existing course assignments.",
    "course.assignment.delete": "Can permanently delete course assignments.",
    "submission.list": "Can view the students who submitted an assignment and their submission summaries.",
    "submission.view": "Can open the full contents of individual assignment submissions.",
    "submission.return": "Can mark submitted assignments as returned to students.",
    "submission.delete": "Can permanently delete returned submissions and their attachments.",
    "submission.grade": "Can add and update scores and feedback on submitted assignments.",
    "submission.removeGrade": "Can remove an existing score and feedback from an assignment submission.",
    "course.student.list": "Can view which students are assigned to courses.",
    "course.student.assign": "Can assign school students to courses.",
    "course.student.remove": "Can remove students from courses.",
    "student.list": "Can view students in this school.",
    "student.view": "Can open individual student profiles and view their full details.",
    "student.create": "Can add new students to this school.",
    "student.update": "Can edit student details and login access.",
    "student.delete": "Can permanently delete students.",
    "log.list": "Can view this school's audit logs.",
}

const permissionCategoryLabels: Record<string, string> = {
    school: "School",
    staffInvitations: "Staff invitations",
    staff: "Staff",
    academicYear: "Academic years",
    grade: "Grades",
    role: "Roles",
    course: "Courses",
    coursePosts: "Course posts",
    postAttachments: "Post attachments",
    courseAssignments: "Course assignments",
    assignmentSubmissions: "Assignment submissions",
    courseStudents: "Course students",
    student: "Students",
    log: "Logs",
}

interface PermissionPreset {
    id: string
    name: string
    description: string
    icon: LucideIcon
    permissions: string[] | "all"
}

const permissionPresets: PermissionPreset[] = [
    {
        id: "administrator",
        name: "Administrator",
        description: "Full access to every available permission.",
        icon: Crown,
        permissions: "all",
    },
    {
        id: "academic-manager",
        name: "Academic manager",
        description: "Manages academic years and grade structure.",
        icon: BookOpen,
        permissions: [
            "school.view", "school.promote", "academicYear.create", "academicYear.list",
            "academicYear.toggleActive", "academicYear.delete", "grade.list", "grade.create",
            "grade.update", "grade.delete", "course.list", "course.create", "course.update", "course.delete",
        ],
    },
    {
        id: "staff-manager",
        name: "Staff manager",
        description: "Manages staff access and assigned roles.",
        icon: UserCog,
        permissions: [
            "school.view", "staff.view", "staff.create", "staff.delete", "staff.role.add",
            "staff.role.remove", "staff.role.list", "school.invite.list", "school.invite.cancel", "role.list",
        ],
    },
    {
        id: "teacher",
        name: "Teacher",
        description: "Views the school, academic years, and grades.",
        icon: Eye,
        permissions: ["school.view", "academicYear.list", "grade.list", "course.list", "course.post.list"],
    },
]

function fallbackPermissionLabel(permission: string) {
    const action = permission.split(".").slice(1).join(" ")
    return action.replace(/([a-z])([A-Z])/g, "$1 $2").replace(/^./, (character) => character.toUpperCase())
}

function fallbackPermissionDescription(label: string) {
    return `Allows this role to ${label.charAt(0).toLowerCase()}${label.slice(1)}.`
}

function permissionCategory(permission: string) {
    if (permission.startsWith("school.invite.")) return "staffInvitations"
    if (permission.startsWith("post.attachment.")) return "postAttachments"
    if (permission.startsWith("course.post.")) return "coursePosts"
    if (permission.startsWith("course.assignment.")) return "courseAssignments"
    if (permission.startsWith("submission.")) return "assignmentSubmissions"
    if (permission.startsWith("course.student.")) return "courseStudents"
    return permission.split(".")[0]
}

function permissionActionPosition(permission: string) {
    const action = permission.split(".").at(-1) ?? ""
    const order = ["view", "list", "create", "add", "assign", "update", "toggleActive", "promote", "return", "remove", "cancel", "delete"]
    const position = order.indexOf(action)
    return position === -1 ? order.length : position
}

function RoleListItem({role, busy, canUpdate, canDelete, canManagePermissions, onDragStart, onDragEnd, onPermissions, onEdit, onDelete}: {
    role: Role
    busy: boolean
    canUpdate: boolean
    canDelete: boolean
    canManagePermissions: boolean
    onDragStart: () => void
    onDragEnd: () => void
    onPermissions: () => void
    onEdit: () => void
    onDelete: () => void
}) {
    const {t} = useLocale()
    const dragControls = useDragControls()
    const [isDragging, setIsDragging] = useState(false)

    return (
        <Reorder.Item
            as="div"
            value={role}
            drag="y"
            dragControls={dragControls}
            dragListener={false}
            dragElastic={0.06}
            dragMomentum={false}
            layout="position"
            transition={{layout: {type: "spring", stiffness: 520, damping: 42}}}
            whileDrag={{scale: 1.01, zIndex: 30, boxShadow: "0 18px 40px rgb(0 0 0 / 0.16)"}}
            onDragStart={() => {
                setIsDragging(true)
                onDragStart()
            }}
            onDragEnd={() => {
                setIsDragging(false)
                onDragEnd()
            }}
            className={`relative list-none border-b bg-card last:overflow-hidden last:rounded-b-xl ${isDragging ? "rounded-xl ring-1 ring-border" : "last:border-b-0"}`}
        >
            <ContextMenu>
                <ContextMenuTrigger className="block">
                    <div className={`group/role flex min-h-16 flex-row items-center gap-3 px-3 py-2.5 transition-colors sm:px-4 ${isDragging ? "rounded-xl bg-card" : "hover:bg-muted/45"}`}>
                <button
                    type="button"
                    aria-label={t("staff.roles.reorderLabel", {name: role.name})}
                    title={t("staff.roles.reorderTitle", {name: role.name})}
                    disabled={busy || !canUpdate}
                    onPointerDown={(event) => {
                        if (!busy && canUpdate) dragControls.start(event)
                    }}
                    className="-m-1 flex size-9 shrink-0 touch-none cursor-grab items-center justify-center rounded-md text-muted-foreground/70 transition-colors hover:bg-muted hover:text-foreground active:cursor-grabbing disabled:cursor-not-allowed disabled:opacity-40"
                >
                    <GripVertical className="size-5" />
                </button>

                <span
                    className="size-5 shrink-0 rounded-full border-2 border-background shadow-sm ring-1 ring-border"
                    style={{backgroundColor: `#${role.color}`}}
                />
                <div className="min-w-0 flex-1">
                    <p className="truncate text-sm font-semibold">{role.name}</p>
                </div>

                {canManagePermissions && <Button className="text-muted-foreground hover:text-foreground" size="sm" variant="ghost" aria-label={t("staff.roles.managePermissionsLabel", {name: role.name})} disabled={busy} onClick={onPermissions}>
                    <KeyRound />
                    <span className="hidden lg:inline">{t("staff.roles.permissions")}</span>
                </Button>}

                {(canUpdate || canDelete) && <div className="flex shrink-0 items-center gap-0.5 transition-opacity sm:opacity-0 sm:group-focus-within/role:opacity-100 sm:group-hover/role:opacity-100">
                    {canUpdate && <Button className="text-muted-foreground hover:text-foreground" size="icon-sm" variant="ghost" aria-label={t("staff.roles.editLabel", {name: role.name})} disabled={busy} onClick={onEdit}>
                        <Pencil />
                    </Button>}
                    {canDelete && <Button className="text-muted-foreground hover:bg-destructive/8 hover:text-destructive" size="icon-sm" variant="ghost" aria-label={t("staff.roles.deleteLabel", {name: role.name})} disabled={busy} onClick={onDelete}>
                        <Trash2 />
                    </Button>}
                </div>}
                    </div>
                </ContextMenuTrigger>
                {(canUpdate || canManagePermissions) && <ContextMenuPopup className="w-52">
                    {canUpdate && <ContextMenuItem disabled={busy} onClick={onEdit}>
                        <Pencil /> {t("staff.roles.edit")}
                    </ContextMenuItem>}
                    {canManagePermissions && <ContextMenuItem disabled={busy} onClick={onPermissions}>
                        <KeyRound /> {t("staff.roles.managePermissions")}
                    </ContextMenuItem>}
                </ContextMenuPopup>}
            </ContextMenu>
        </Reorder.Item>
    )
}

export default function RolesClientPage({schoolID, roles, availablePermissions, access = emptySchoolAccess}: {
    schoolID: string
    roles: Role[]
    availablePermissions: string[]
    access: SchoolAccess
}) {
    const router = useRouter()
    const {locale, t} = useLocale()
    const initialRoles = [...roles].sort((a, b) => b.position - a.position)
    const actorPosition = highestRolePosition(access)
    const maximumCreatePosition = access.owner
        ? roles.length
        : Math.min(roles.length, actorPosition - 1)
    const canCreateAtAnyPosition = access.owner || maximumCreatePosition >= 0
    const [orderedRoles, setOrderedRoles] = useState(initialRoles)
    const latestOrderRef = useRef(initialRoles)
    const dragStartOrderRef = useRef(initialRoles)
    const draggedRoleIDRef = useRef<string | null>(null)
    const reorderSavingRef = useRef(false)
    const [reorderSaving, setReorderSaving] = useState(false)
    const canCreateRoles = hasSchoolPermission(access, "role.create") && canCreateAtAnyPosition
    const canUpdateRoles = hasSchoolPermission(access, "role.update")
    const canDeleteRoles = hasSchoolPermission(access, "role.delete")
    const canUpdatePermissions = hasSchoolPermission(access, "role.permission.update")

    const [createOpen, setCreateOpen] = useState(false)
    const [creating, setCreating] = useState(false)
    const [name, setName] = useState("")
    const [position, setPosition] = useState<number | null>(Math.max(0, maximumCreatePosition))
    const [color, setColor] = useState("6366F1")

    const [editTarget, setEditTarget] = useState<Role | null>(null)
    const [editName, setEditName] = useState("")
    const [editPosition, setEditPosition] = useState<number | null>(null)
    const [editColor, setEditColor] = useState("")
    const [updating, setUpdating] = useState(false)

    const [deleteTarget, setDeleteTarget] = useState<Role | null>(null)
    const [deleteConfirmation, setDeleteConfirmation] = useState("")
    const [deleting, setDeleting] = useState(false)

    const [permissionRole, setPermissionRole] = useState<Role | null>(null)
    const [enabledPermissions, setEnabledPermissions] = useState<Set<string>>(new Set())
    const [pendingPermissions, setPendingPermissions] = useState<Set<string>>(new Set())
    const [applyingPreset, setApplyingPreset] = useState<string | null>(null)
    const [presetProgress, setPresetProgress] = useState<{completed: number; total: number} | null>(null)

    async function createRole(event: React.SubmitEvent<HTMLFormElement>) {
        event.preventDefault()
        setCreating(true)
        const res = await handleCreateRole(schoolID, {name, position: position ?? Number.NaN, color})
        setCreating(false)
        if (!res.ok) return toast.error(t("staff.roles.error.action"))

        setName("")
        setPosition(Math.max(0, maximumCreatePosition))
        setColor("6366F1")
        setCreateOpen(false)
        toast.success(t("staff.roles.created"))
        router.refresh()
    }

    function openEdit(role: Role) {
        setEditTarget(role)
        setEditName(role.name)
        setEditPosition(role.position)
        setEditColor(role.color.toUpperCase())
    }

    function openPermissions(role: Role) {
        setPermissionRole(role)
        setEnabledPermissions(new Set(role.permissions ?? []))
        setPendingPermissions(new Set())
        setApplyingPreset(null)
        setPresetProgress(null)
    }

    async function setRolePermission(permission: string, allow: boolean) {
        if (!permissionRole || applyingPreset || pendingPermissions.has(permission)) return
        const targetRole = permissionRole

        setEnabledPermissions((current) => {
            const next = new Set(current)
            if (allow) next.add(permission)
            else next.delete(permission)
            return next
        })
        setPendingPermissions((current) => new Set(current).add(permission))

        const res = await handleSetRolePermission(targetRole.id, permission, allow)
        setPendingPermissions((current) => {
            const next = new Set(current)
            next.delete(permission)
            return next
        })

        if (!res.ok) {
            setEnabledPermissions((current) => {
                const next = new Set(current)
                if (allow) next.delete(permission)
                else next.add(permission)
                return next
            })
            return toast.error(t("staff.roles.error.action"))
        }

        setOrderedRoles((currentRoles) => {
            const updatedRoles = currentRoles.map((role) => {
                if (role.id !== targetRole.id) return role
                const permissions = new Set(role.permissions ?? [])
                if (allow) permissions.add(permission)
                else permissions.delete(permission)
                return {...role, permissions: [...permissions].sort()}
            })
            latestOrderRef.current = updatedRoles
            return updatedRoles
        })
    }

    async function applyPermissionPreset(preset: PermissionPreset) {
        if (!permissionRole || applyingPreset) return
        const targetRole = permissionRole
        const originalPermissions = new Set(enabledPermissions)
        const desiredPermissions = new Set(
            (preset.permissions === "all" ? availablePermissions : preset.permissions)
                .filter((permission) => availablePermissions.includes(permission)),
        )
        const changes = availablePermissions
            .filter((permission) => originalPermissions.has(permission) !== desiredPermissions.has(permission))
            .map((permission) => ({permission, allow: desiredPermissions.has(permission)}))

        if (!changes.length) {
            toast.success(t("staff.roles.presetAlreadyApplied", {name: localizedPermissionPreset(locale, preset.id, preset.name, preset.description).name}))
            return
        }

        setApplyingPreset(preset.id)
        setPresetProgress({completed: 0, total: changes.length})
        setPendingPermissions(new Set(changes.map(({permission}) => permission)))
        setEnabledPermissions(desiredPermissions)

        const results = await Promise.all(changes.map(async (change) => {
            const res = await handleSetRolePermission(targetRole.id, change.permission, change.allow)
            setPresetProgress((current) => current
                ? {...current, completed: current.completed + 1}
                : current
            )
            return {...change, ok: res.ok}
        }))

        const finalPermissions = new Set(originalPermissions)
        for (const result of results) {
            if (!result.ok) continue
            if (result.allow) finalPermissions.add(result.permission)
            else finalPermissions.delete(result.permission)
        }

        setEnabledPermissions(finalPermissions)
        setOrderedRoles((currentRoles) => {
            const updatedRoles = currentRoles.map((role) => role.id === targetRole.id
                ? {...role, permissions: [...finalPermissions].sort()}
                : role
            )
            latestOrderRef.current = updatedRoles
            return updatedRoles
        })
        setPendingPermissions(new Set())
        setApplyingPreset(null)
        setPresetProgress(null)

        const failed = results.filter((result) => !result.ok).length
        if (failed) {
            toast.error(t("staff.roles.presetPartial", {name: localizedPermissionPreset(locale, preset.id, preset.name, preset.description).name, count: failed}))
        } else {
            toast.success(t("staff.roles.presetApplied", {name: localizedPermissionPreset(locale, preset.id, preset.name, preset.description).name}))
        }
    }

    async function updateRole(event: React.SubmitEvent<HTMLFormElement>) {
        event.preventDefault()
        if (!editTarget) return

        setUpdating(true)
        const res = await handleUpdateRole(editTarget.id, {
            name: editName,
            position: editPosition ?? Number.NaN,
            color: editColor,
        })
        setUpdating(false)
        if (!res.ok) return toast.error(t("staff.roles.error.action"))

        const updatedRoles = orderedRoles
            .map((role) => role.id === editTarget.id
                ? {...role, name: editName.trim(), position: editPosition!, color: editColor.toUpperCase()}
                : role)
            .sort((a, b) => b.position - a.position)
        latestOrderRef.current = updatedRoles
        setOrderedRoles(updatedRoles)
        setEditTarget(null)
        toast.success(t("staff.roles.updated"))
        router.refresh()
    }

    async function deleteRole() {
        if (!deleteTarget) return
        if (deleteConfirmation !== deleteTarget.name) return

        setDeleting(true)
        const res = await handleDeleteRole(deleteTarget.id)
        setDeleting(false)
        if (!res.ok) return toast.error(t("staff.roles.error.action"))

        const remainingRoles = orderedRoles.filter((role) => role.id !== deleteTarget.id)
        latestOrderRef.current = remainingRoles
        setOrderedRoles(remainingRoles)
        setDeleteTarget(null)
        setDeleteConfirmation("")
        toast.success(t("staff.roles.deleted"))
        router.refresh()
    }

    function reorderRoles(nextOrder: Role[]) {
        if (reorderSavingRef.current) return

        const protectedRoleIDs = new Set(initialRoles
            .filter((role) => !canManagePosition(access, role.position))
            .map((role) => role.id)
        )
        const protectedRoles = initialRoles.filter((role) => protectedRoleIDs.has(role.id))
        const movableRoles = nextOrder.filter((role) => !protectedRoleIDs.has(role.id))
        const allowedOrder = [...protectedRoles, ...movableRoles]

        latestOrderRef.current = allowedOrder
        setOrderedRoles(allowedOrder)
    }

    function startDragging(roleID: string) {
        dragStartOrderRef.current = orderedRoles
        draggedRoleIDRef.current = roleID
    }

    async function finishDragging() {
        if (reorderSavingRef.current) return

        const previousOrder = dragStartOrderRef.current
        const nextOrder = latestOrderRef.current
        const draggedRoleID = draggedRoleIDRef.current
        if (previousOrder.map((role) => role.id).join() === nextOrder.map((role) => role.id).join()) return
        if (!draggedRoleID) return

        reorderSavingRef.current = true
        setReorderSaving(true)
        const res = await handleReorderRoles(
            previousOrder.map(({id, name, color, position}) => ({id, name, color, position})),
            nextOrder.map((role) => role.id),
            draggedRoleID,
        )
        reorderSavingRef.current = false
        setReorderSaving(false)

        const positions = "positions" in res ? res.positions : undefined
        if (!res.ok || !positions) {
            latestOrderRef.current = previousOrder
            setOrderedRoles(previousOrder)
            router.refresh()
            return toast.error(t("staff.roles.orderError"))
        }

        const savedOrder = nextOrder.map((role) => ({...role, position: positions[role.id] || 0}))
        latestOrderRef.current = savedOrder
        setOrderedRoles(savedOrder)
        toast.success(t("staff.roles.orderSaved"))
        router.refresh()
    }

    const normalizedColor = /^[0-9A-Fa-f]{6}$/.test(color) ? color : "000000"
    const roleAbove = position === null
        ? undefined
        : orderedRoles.filter((role) => role.position >= position).at(-1)
    const roleBelow = position === null
        ? undefined
        : orderedRoles.find((role) => role.position < position)
    const normalizedEditColor = /^[0-9A-Fa-f]{6}$/.test(editColor) ? editColor : "000000"
    const permissionGroups = Object.entries(availablePermissions.reduce<Record<string, string[]>>((groups, permission) => {
        const category = permissionCategory(permission)
        groups[category] ??= []
        groups[category].push(permission)
        return groups
    }, {})).map(([category, permissions]) => [
        category,
        permissions.sort((first, second) => permissionActionPosition(first) - permissionActionPosition(second)),
    ] as [string, string[]]).sort(([first], [second]) => {
        const order = ["school", "staffInvitations", "staff", "student", "academicYear", "grade", "course", "coursePosts", "role", "log"]
        return order.indexOf(first) - order.indexOf(second)
    })

    return (
        <div className="space-y-6">
            <div className="flex flex-wrap items-end justify-between gap-3">
                <div>
                    <PageTitle centered={false}>{t("staff.roles.title")}</PageTitle>
                    <p className="text-muted-foreground">
                        {t(reorderSaving ? "staff.roles.savingOrder" : canUpdateRoles ? "staff.roles.dragDescription" : "staff.roles.orderDescription")}
                    </p>
                </div>
                {canCreateRoles && <Button onClick={() => setCreateOpen(true)} disabled={reorderSaving}><Plus /> {t("staff.roles.create")}</Button>}
            </div>

            {orderedRoles.length ? (
                <Card className="rounded-xl">
                    <div className="flex items-center justify-between gap-4 rounded-t-xl border-b bg-muted/35 px-4 py-3.5">
                        <div className="flex min-w-0 items-center gap-3">
                            <div className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-indigo-500/10 text-indigo-600 dark:text-indigo-400">
                                <ShieldCheck className="size-4.5" />
                            </div>
                            <div className="min-w-0">
                                <p className="truncate text-sm font-semibold">{t("staff.roles.hierarchy")}</p>
                                <p className="truncate text-xs text-muted-foreground">{t("staff.roles.hierarchyDescription")}</p>
                            </div>
                        </div>
                        <div className="flex shrink-0 items-center gap-1.5">
                            <Dialog>
                                <DialogTrigger
                                    render={<Button size="icon-sm" variant="ghost" />}
                                    aria-label={t("staff.roles.hierarchyHelpOpen")}
                                    title={t("staff.roles.hierarchyHelpOpen")}
                                >
                                    <CircleHelp />
                                </DialogTrigger>
                                <DialogPopup className="sm:max-w-lg">
                                    <DialogHeader>
                                        <DialogTitle>{t("staff.roles.hierarchyHelpTitle")}</DialogTitle>
                                        <DialogDescription>{t("staff.roles.hierarchyHelpDescription")}</DialogDescription>
                                    </DialogHeader>
                                    <DialogPanel className="space-y-4">
                                        <ol className="space-y-3">
                                            {([
                                                ["staff.roles.hierarchyHelpOrderTitle", "staff.roles.hierarchyHelpOrderDescription"],
                                                ["staff.roles.hierarchyHelpHighestTitle", "staff.roles.hierarchyHelpHighestDescription"],
                                                ["staff.roles.hierarchyHelpManageTitle", "staff.roles.hierarchyHelpManageDescription"],
                                            ] as const).map(([title, description], index) => (
                                                <li key={title} className="flex gap-3 rounded-xl border bg-muted/25 p-3.5">
                                                    <span className="flex size-7 shrink-0 items-center justify-center rounded-full bg-primary/10 text-sm font-semibold text-primary">
                                                        {index + 1}
                                                    </span>
                                                    <div className="min-w-0 pt-0.5">
                                                        <p className="text-sm font-semibold">{t(title)}</p>
                                                        <p className="mt-1 text-sm leading-relaxed text-muted-foreground">{t(description)}</p>
                                                    </div>
                                                </li>
                                            ))}
                                        </ol>
                                        <div className="rounded-xl border border-indigo-500/20 bg-indigo-500/7 p-3.5">
                                            <p className="text-sm font-semibold">{t("staff.roles.hierarchyHelpPermissionsTitle")}</p>
                                            <p className="mt-1 text-sm leading-relaxed text-muted-foreground">{t("staff.roles.hierarchyHelpPermissionsDescription")}</p>
                                        </div>
                                        <p className="text-xs leading-relaxed text-muted-foreground">{t("staff.roles.hierarchyHelpOwnerNote")}</p>
                                    </DialogPanel>
                                    <DialogFooter>
                                        <DialogClose render={<Button />}>{t("staff.roles.hierarchyHelpClose")}</DialogClose>
                                    </DialogFooter>
                                </DialogPopup>
                            </Dialog>
                            <Badge variant="secondary">{t(orderedRoles.length === 1 ? "staff.roles.roleOne" : "staff.roles.roleOther", {count: orderedRoles.length})}</Badge>
                        </div>
                    </div>
                    <Reorder.Group as="div" axis="y" values={orderedRoles} onReorder={reorderRoles} className="rounded-b-xl">
                        {orderedRoles.map((role) => (
                            <RoleListItem
                                key={role.id}
                                role={role}
                                busy={reorderSaving}
                                canUpdate={canUpdateRoles && canManagePosition(access, role.position)}
                                canDelete={canDeleteRoles && canManagePosition(access, role.position)}
                                canManagePermissions={canUpdatePermissions && canManagePosition(access, role.position)}
                                onDragStart={() => startDragging(role.id)}
                                onDragEnd={finishDragging}
                                onPermissions={() => openPermissions(role)}
                                onEdit={() => openEdit(role)}
                                onDelete={() => setDeleteTarget(role)}
                            />
                        ))}
                    </Reorder.Group>
                </Card>
            ) : (
                <Card>
                    <Empty>
                        <EmptyHeader>
                            <EmptyMedia variant="icon"><ShieldCheck /></EmptyMedia>
                            <EmptyTitle>{t("staff.roles.empty")}</EmptyTitle>
                            <EmptyDescription>{t("staff.roles.emptyDescription")}</EmptyDescription>
                        </EmptyHeader>
                    </Empty>
                </Card>
            )}

            <Dialog open={createOpen} onOpenChange={(open) => {
                if (!creating) setCreateOpen(open)
            }}>
                <DialogPopup className="sm:max-w-lg">
                    <DialogHeader>
                        <DialogTitle>{t("staff.roles.create")}</DialogTitle>
                        <DialogDescription>{t("staff.roles.createDescription")}</DialogDescription>
                    </DialogHeader>
                    <Form className="contents" onSubmit={createRole}>
                        <DialogPanel className="space-y-5">
                            <Field>
                                <FieldLabel>{t("staff.roles.name")}</FieldLabel>
                                <Input value={name} onChange={(event) => setName(event.target.value)} maxLength={32} placeholder={t("staff.roles.namePlaceholder")} required autoFocus />
                            </Field>
                            <Field>
                                <FieldLabel>{t("staff.roles.color")}</FieldLabel>
                                <div className="flex items-center gap-2">
                                    <Input className="h-9 w-14 cursor-pointer p-1" type="color" value={`#${normalizedColor}`} onChange={(event) => setColor(event.target.value.slice(1).toUpperCase())} />
                                    <div className="relative flex-1">
                                        <span className="pointer-events-none absolute inset-y-0 left-3 flex items-center text-muted-foreground">#</span>
                                        <Input className="pl-7 font-mono uppercase" value={color} onChange={(event) => setColor(event.target.value.replace(/^#/, "").slice(0, 6))} pattern="[0-9A-Fa-f]{6}" maxLength={6} required />
                                    </div>
                                </div>
                            </Field>
                            <Field>
                                <FieldLabel>{t("staff.roles.position")}</FieldLabel>
                                <NumberField min={0} max={Math.max(0, maximumCreatePosition)} step={1} required value={position} onValueChange={setPosition}>
                                    <NumberFieldGroup>
                                        <NumberFieldDecrement aria-label={t("staff.roles.decreasePosition")} />
                                        <NumberFieldInput />
                                        <NumberFieldIncrement aria-label={t("staff.roles.increasePosition")} />
                                    </NumberFieldGroup>
                                </NumberField>
                                <FieldDescription>{t("staff.roles.positionHelp")}</FieldDescription>
                            </Field>

                            <div className="space-y-2">
                                <p className="text-sm font-medium">{t("staff.roles.preview")}</p>
                                <div className="overflow-hidden rounded-lg border bg-muted/30 p-2">
                                    {roleAbove && <RolePreviewRow role={{...roleAbove, position: roleAbove.position + 1}} />}
                                    <div className="flex items-center justify-between gap-3 rounded-md border bg-background px-3 py-2.5 shadow-xs">
                                        <div className="flex min-w-0 items-center gap-2">
                                            <span className="size-3 shrink-0 rounded-full" style={{backgroundColor: `#${normalizedColor}`}} />
                                            <span className="truncate text-sm font-medium">{name.trim() || t("staff.roles.newRole")}</span>
                                            <Badge variant="secondary">{t("staff.roles.new")}</Badge>
                                        </div>
                                        <span className="text-sm tabular-nums text-muted-foreground">{position ?? "—"}</span>
                                    </div>
                                    {roleBelow && <RolePreviewRow role={roleBelow} />}
                                </div>
                            </div>
                        </DialogPanel>
                        <DialogFooter>
                            <DialogClose render={<Button variant="outline" disabled={creating} />}>{t("staff.roles.cancel")}</DialogClose>
                            <Button type="submit" loading={creating} disabled={!name.trim() || position === null || !/^[0-9A-Fa-f]{6}$/.test(color)}>{t("staff.roles.create")}</Button>
                        </DialogFooter>
                    </Form>
                </DialogPopup>
            </Dialog>

            <Dialog open={permissionRole !== null} onOpenChange={(open) => {
                if (!open && pendingPermissions.size === 0 && !applyingPreset) setPermissionRole(null)
            }}>
                <DialogPopup className="sm:max-w-2xl">
                    <DialogHeader>
                        <div className="flex items-start gap-3 pr-8">
                            <span
                                className="mt-0.5 size-5 shrink-0 rounded-full border-2 border-background shadow-sm ring-1 ring-border"
                                style={{backgroundColor: `#${permissionRole?.color ?? "000000"}`}}
                            />
                            <div className="min-w-0">
                                <DialogTitle>{t("staff.roles.manageTitle", {name: permissionRole?.name ?? ""})}</DialogTitle>
                                <DialogDescription>{t("staff.roles.changesImmediate")}</DialogDescription>
                            </div>
                        </div>
                    </DialogHeader>
                    <DialogPanel className="max-h-[65vh] space-y-5 overflow-y-auto">
                        <section className="space-y-2.5">
                            <div>
                                <h3 className="text-sm font-semibold">{t("staff.roles.presets")}</h3>
                                <p className="text-xs text-muted-foreground">{t("staff.roles.presetsDescription")}</p>
                            </div>
                            <div className="grid gap-2 sm:grid-cols-2">
                                {permissionPresets.map((preset) => {
                                    const Icon = preset.icon
                                    const localizedPreset = localizedPermissionPreset(locale, preset.id, preset.name, preset.description)
                                    return (
                                        <Button
                                            key={preset.id}
                                            variant="outline"
                                            className="h-auto min-h-16 justify-start whitespace-normal px-3 py-2.5 text-left"
                                            loading={applyingPreset === preset.id}
                                            disabled={applyingPreset !== null || pendingPermissions.size > 0}
                                            onClick={() => applyPermissionPreset(preset)}
                                        >
                                            <Icon className="self-start" />
                                            <span className="min-w-0">
                                                <span className="block font-medium">{localizedPreset.name}</span>
                                                <span className="block text-xs font-normal text-muted-foreground">{localizedPreset.description}</span>
                                            </span>
                                        </Button>
                                    )
                                })}
                            </div>
                            {presetProgress && (
                                <div className="flex items-center justify-between rounded-lg bg-primary/8 px-3 py-2 text-xs text-primary">
                                    <span>{t("staff.roles.applyingPreset")}</span>
                                    <span className="font-medium tabular-nums">{presetProgress.completed}/{presetProgress.total}</span>
                                </div>
                            )}
                        </section>

                        <div className="flex items-center justify-between rounded-lg bg-muted/45 px-3 py-2.5 text-sm">
                            <span className="text-muted-foreground">{t("staff.roles.enabledPermissions")}</span>
                            <Badge variant="secondary">
                                {t("staff.roles.enabledCount", {enabled: availablePermissions.filter((permission) => enabledPermissions.has(permission)).length, total: availablePermissions.length})}
                            </Badge>
                        </div>

                        {permissionGroups.map(([category, permissions]) => (
                            <section key={category} className="space-y-2">
                                <h3 className="px-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                                    {localizedPermissionCategory(locale, category, permissionCategoryLabels[category] ?? fallbackPermissionLabel(`${category}.permissions`))}
                                </h3>
                                <div className="divide-y overflow-hidden rounded-xl border">
                                    {permissions.map((permission) => {
                                        const label = localizedPermissionLabel(locale, permission, permissionLabels[permission] ?? fallbackPermissionLabel(permission))
                                        const description = localizedPermissionDescription(locale, label, permissionDescriptions[permission] ?? fallbackPermissionDescription(label))
                                        const switchID = `permission-${permissionRole?.id}-${permission}`
                                        const pending = pendingPermissions.has(permission)
                                        const destructive = permission.endsWith(".delete")
                                        return (
                                            <div key={permission} className={`flex items-center gap-4 px-4 py-3 transition-colors ${destructive ? "bg-destructive/4 hover:bg-destructive/7" : "bg-card hover:bg-muted/30"}`}>
                                                <label htmlFor={switchID} className="min-w-0 flex-1 cursor-pointer">
                                                    <span className={`block text-sm font-medium ${destructive ? "text-destructive" : ""}`}>{label}</span>
                                                    <span className={`mt-0.5 block text-xs leading-relaxed ${destructive ? "text-destructive/70" : "text-muted-foreground"}`}>{description}</span>
                                                </label>
                                                <Switch
                                                    id={switchID}
                                                    className={destructive ? "data-checked:bg-destructive" : undefined}
                                                    checked={enabledPermissions.has(permission)}
                                                    disabled={pending || applyingPreset !== null}
                                                    aria-label={t(enabledPermissions.has(permission) ? "staff.roles.disablePermission" : "staff.roles.enablePermission", {label})}
                                                    onCheckedChange={(checked) => setRolePermission(permission, checked)}
                                                />
                                            </div>
                                        )
                                    })}
                                </div>
                            </section>
                        ))}
                    </DialogPanel>
                </DialogPopup>
            </Dialog>

            <Dialog open={editTarget !== null} onOpenChange={(open) => {
                if (!open && !updating) setEditTarget(null)
            }}>
                <DialogPopup className="sm:max-w-md">
                    <DialogHeader>
                        <DialogTitle>{t("staff.roles.updateTitle")}</DialogTitle>
                        <DialogDescription>{t("staff.roles.updateDescription")}</DialogDescription>
                    </DialogHeader>
                    <Form className="contents" onSubmit={updateRole}>
                        <DialogPanel className="space-y-4">
                            <Field>
                                <FieldLabel>{t("staff.roles.name")}</FieldLabel>
                                <Input value={editName} onChange={(event) => setEditName(event.target.value)} maxLength={32} required autoFocus />
                            </Field>
                            <Field>
                                <FieldLabel>{t("staff.roles.color")}</FieldLabel>
                                <div className="flex items-center gap-2">
                                    <Input className="h-9 w-14 cursor-pointer p-1" type="color" value={`#${normalizedEditColor}`} onChange={(event) => setEditColor(event.target.value.slice(1).toUpperCase())} />
                                    <div className="relative flex-1">
                                        <span className="pointer-events-none absolute inset-y-0 left-3 flex items-center text-muted-foreground">#</span>
                                        <Input className="pl-7 font-mono uppercase" value={editColor} onChange={(event) => setEditColor(event.target.value.replace(/^#/, "").slice(0, 6))} pattern="[0-9A-Fa-f]{6}" maxLength={6} required />
                                    </div>
                                </div>
                            </Field>
                            <Field>
                                <FieldLabel>{t("staff.roles.position")}</FieldLabel>
                                <NumberField
                                    min={0}
                                    max={Math.max(0, access.owner ? orderedRoles.length - 1 : Math.min(orderedRoles.length - 1, actorPosition - 1))}
                                    step={1}
                                    required
                                    value={editPosition}
                                    onValueChange={setEditPosition}
                                >
                                    <NumberFieldGroup>
                                        <NumberFieldDecrement aria-label={t("staff.roles.decreasePosition")} />
                                        <NumberFieldInput />
                                        <NumberFieldIncrement aria-label={t("staff.roles.increasePosition")} />
                                    </NumberFieldGroup>
                                </NumberField>
                                <FieldDescription>{t("staff.roles.positionHelp")}</FieldDescription>
                            </Field>
                        </DialogPanel>
                        <DialogFooter>
                            <DialogClose render={<Button variant="outline" disabled={updating} />}>{t("staff.roles.cancel")}</DialogClose>
                            <Button type="submit" loading={updating} disabled={!editName.trim() || editPosition === null || !/^[0-9A-Fa-f]{6}$/.test(editColor)}>{t("staff.roles.save")}</Button>
                        </DialogFooter>
                    </Form>
                </DialogPopup>
            </Dialog>

            <AlertDialog open={deleteTarget !== null} onOpenChange={(open) => {
                if (!open && !deleting) {
                    setDeleteTarget(null)
                    setDeleteConfirmation("")
                }
            }}>
                <AlertDialogPopup className="border-destructive/30 sm:max-w-lg">
                    <AlertDialogHeader>
                        <div className="mb-2 flex size-12 items-center justify-center rounded-2xl bg-destructive/10 text-destructive">
                            <TriangleAlert className="size-6" />
                        </div>
                        <AlertDialogTitle className="text-destructive">{t("staff.roles.deleteTitle")}</AlertDialogTitle>
                        <AlertDialogDescription className="space-y-3 text-left">
                            <span className="block">{t("staff.roles.deleteDescription", {name: deleteTarget?.name ?? ""})}</span>
                            <span className="block font-medium text-destructive">{t("staff.roles.deleteIrreversible")}</span>
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <Field className="px-6 pb-2">
                        <FieldLabel htmlFor="role-delete-confirmation">{t("staff.roles.typeName")}</FieldLabel>
                        <Input
                            id="role-delete-confirmation"
                            value={deleteConfirmation}
                            onChange={(event) => setDeleteConfirmation(event.target.value)}
                            placeholder={deleteTarget?.name}
                            autoComplete="off"
                            disabled={deleting}
                            autoFocus
                        />
                    </Field>
                    <AlertDialogFooter>
                        <AlertDialogClose render={<Button variant="outline" disabled={deleting} />}>{t("staff.roles.cancel")}</AlertDialogClose>
                        <Button
                            variant="destructive"
                            loading={deleting}
                            disabled={!deleteTarget || deleteConfirmation !== deleteTarget.name || deleting}
                            onClick={deleteRole}
                        >
                            {t("staff.roles.deletePermanently")}
                        </Button>
                    </AlertDialogFooter>
                </AlertDialogPopup>
            </AlertDialog>
        </div>
    )
}

function RolePreviewRow({role}: {role: Role}) {
    return (
        <div className="flex items-center justify-between gap-3 px-3 py-2 text-sm text-muted-foreground">
            <div className="flex min-w-0 items-center gap-2">
                <span className="size-2.5 shrink-0 rounded-full" style={{backgroundColor: `#${role.color}`}} />
                <span className="truncate">{role.name}</span>
            </div>
            <span className="tabular-nums">{role.position}</span>
        </div>
    )
}
