import assert from "node:assert/strict"
import test from "node:test"
import {
    buildAssignmentUpdateInput,
    getAssignmentArchiveAction,
    splitAssignmentsByArchiveState,
} from "./staff_assignment.ts"

const assignment = {
    id: "82461601281605633",
    referencedPost: {id: "82972987971600385", title: "Instructions"},
    title: "Worksheet",
    description: "Complete every question.",
    dueDate: "2026-10-12T12:00:00.000Z",
    submissionsEnabled: true,
    submissionsCloseAt: "2026-10-13T12:00:00.000Z",
    createdAt: "2026-10-08T12:00:00.000Z",
    archivedAt: null,
}

test("archive updates preserve assignment fields and add the requested archived state", () => {
    assert.deepEqual(buildAssignmentUpdateInput(assignment, true), {
        referencedPostId: "82972987971600385",
        title: "Worksheet",
        description: "Complete every question.",
        dueDate: "2026-10-12T12:00:00.000Z",
        submissionsEnabled: true,
        submissionsCloseAt: "2026-10-13T12:00:00.000Z",
        archived: true,
    })
})

test("unarchive updates work for assignments without a referenced post or submissions", () => {
    assert.deepEqual(buildAssignmentUpdateInput({
        ...assignment,
        referencedPost: null,
        submissionsEnabled: false,
        submissionsCloseAt: "2026-10-13T12:00:00.000Z",
        archivedAt: "2026-10-09T12:00:00.000Z",
    }, false), {
        referencedPostId: null,
        title: "Worksheet",
        description: "Complete every question.",
        dueDate: "2026-10-12T12:00:00.000Z",
        submissionsEnabled: false,
        submissionsCloseAt: null,
        archived: false,
    })
})

test("assignment lists separate active and archived records by archivedAt", () => {
    const archived = {
        ...assignment,
        id: "2",
        title: "Old worksheet",
        archivedAt: "2026-10-09T12:00:00.000Z",
    }

    assert.deepEqual(splitAssignmentsByArchiveState([assignment, archived]), {
        active: [assignment],
        archived: [archived],
    })
})

test("active assignments require confirmation before archiving", () => {
    assert.deepEqual(getAssignmentArchiveAction(null), {
        nextArchived: true,
        actionLabel: "Archive assignment",
        title: "Archive assignment?",
        description: "Students won't be able to access or interact with this assignment while it is archived. You can unarchive it at any time.",
    })
})

test("archived assignments require confirmation before restoring student access", () => {
    assert.deepEqual(getAssignmentArchiveAction("2026-10-09T12:00:00.000Z"), {
        nextArchived: false,
        actionLabel: "Unarchive assignment",
        title: "Unarchive assignment?",
        description: "Students will be able to access and interact with this assignment again.",
    })
})
