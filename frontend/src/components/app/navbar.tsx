interface Props {
    schoolId: number
}

export default function AppNavbar({ schoolId }: Props) {

    return (
        <div>
            <p>{schoolId}</p>
        </div>
    )
}