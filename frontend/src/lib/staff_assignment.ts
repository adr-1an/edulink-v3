export interface StaffAssignmentListItem {
    id: string
    referencedPost: {id: string} | null
    title: string
    description: string
    dueDate: string | null
    submissionsEnabled: boolean
    submissionsCloseAt: string | null
    createdAt: string
    archivedAt: string | null
}

export interface StaffAssignmentUpdateInput {
    referencedPostId: string | null
    title: string
    description: string
    dueDate: string | null
    submissionsEnabled: boolean
    submissionsCloseAt: string | null
    archived: boolean
}

export function buildAssignmentUpdateInput(
    assignment: StaffAssignmentListItem,
    archived: boolean,
): StaffAssignmentUpdateInput {
    return {
        referencedPostId: assignment.referencedPost?.id ?? null,
        title: assignment.title,
        description: assignment.description,
        dueDate: assignment.dueDate,
        submissionsEnabled: assignment.submissionsEnabled,
        submissionsCloseAt: assignment.submissionsEnabled ? assignment.submissionsCloseAt : null,
        archived,
    }
}

export function splitAssignmentsByArchiveState<T extends {archivedAt: string | null}>(
    assignments: T[],
): {active: T[]; archived: T[]} {
    const active: T[] = []
    const archived: T[] = []

    for (const assignment of assignments) {
        if (assignment.archivedAt === null) active.push(assignment)
        else archived.push(assignment)
    }

    return {active, archived}
}

export function getAssignmentArchiveAction(archivedAt: string | null) {
    if (archivedAt === null) {
        return {
            nextArchived: true,
            actionLabel: "Archive assignment",
            title: "Archive assignment?",
            description: "Students won't be able to access or interact with this assignment while it is archived. You can unarchive it at any time.",
        }
    }

    return {
        nextArchived: false,
        actionLabel: "Unarchive assignment",
        title: "Unarchive assignment?",
        description: "Students will be able to access and interact with this assignment again.",
    }
}
