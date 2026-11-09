import {notFound, redirect} from "next/navigation"

export default async function Page({params}: {params: Promise<{studentID: string}>}) {
    const {studentID} = await params
    if (!/^\d+$/.test(studentID)) notFound()
    redirect(`/app/staff/students/${studentID}/guardians`)
}
