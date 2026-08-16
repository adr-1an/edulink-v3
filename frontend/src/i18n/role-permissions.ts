import {type Locale} from "./config"

const polishPermissionLabels: Record<string, string> = {
    "school.view": "Wyświetlanie szkoły",
    "school.update": "Edycja ustawień szkoły",
    "school.delete": "Usuwanie szkoły",
    "school.promote": "Promowanie uczniów",
    "school.invite.list": "Wyświetlanie zaproszeń do personelu",
    "school.invite.cancel": "Anulowanie zaproszeń do personelu",
    "staff.view": "Wyświetlanie personelu",
    "staff.create": "Zapraszanie personelu",
    "staff.delete": "Usuwanie personelu",
    "staff.role.add": "Przypisywanie ról personelu",
    "staff.role.remove": "Odbieranie ról personelu",
    "staff.role.list": "Wyświetlanie ról personelu",
    "academicYear.create": "Tworzenie lat szkolnych",
    "academicYear.list": "Wyświetlanie lat szkolnych",
    "academicYear.toggleActive": "Zmiana aktywnego roku szkolnego",
    "academicYear.delete": "Usuwanie lat szkolnych",
    "grade.list": "Wyświetlanie klas",
    "grade.create": "Tworzenie klas",
    "grade.update": "Edycja klas",
    "grade.delete": "Usuwanie klas",
    "role.create": "Tworzenie ról",
    "role.list": "Wyświetlanie ról",
    "role.update": "Edycja ról",
    "role.delete": "Usuwanie ról",
    "role.permission.update": "Zarządzanie uprawnieniami ról",
    "course.create": "Tworzenie kursów",
    "course.list": "Wyświetlanie kursów",
    "course.update": "Edycja kursów",
    "course.delete": "Usuwanie kursów",
    "course.post.create": "Tworzenie wpisów kursu",
    "course.post.list": "Wyświetlanie wpisów kursu",
    "course.post.view": "Otwieranie wpisów kursu",
    "course.post.update": "Edycja wpisów kursu",
    "course.post.delete": "Usuwanie wpisów kursu",
    "post.attachment.create": "Przesyłanie załączników do wpisów",
    "post.attachment.delete": "Usuwanie załączników z wpisów",
    "course.assignment.create": "Tworzenie zadań kursu",
    "course.assignment.list": "Wyświetlanie zadań kursu",
    "course.assignment.update": "Edycja zadań kursu",
    "course.assignment.delete": "Usuwanie zadań kursu",
    "submission.list": "Wyświetlanie przesłanych zadań",
    "submission.view": "Otwieranie przesłanych zadań",
    "submission.return": "Zwracanie przesłanych zadań",
    "submission.delete": "Usuwanie zwróconych zadań",
    "submission.grade": "Ocenianie przesłanych zadań",
    "submission.removeGrade": "Usuwanie ocen zadań",
    "course.student.list": "Wyświetlanie uczniów kursu",
    "course.student.assign": "Przypisywanie uczniów do kursu",
    "course.student.remove": "Usuwanie uczniów z kursu",
    "student.list": "Wyświetlanie uczniów",
    "student.view": "Otwieranie profili uczniów",
    "student.create": "Tworzenie uczniów",
    "student.update": "Edycja uczniów",
    "student.delete": "Usuwanie uczniów",
    "log.list": "Wyświetlanie dziennika audytu",
}

const polishCategories: Record<string, string> = {
    school: "Szkoła",
    staffInvitations: "Zaproszenia do personelu",
    staff: "Personel",
    academicYear: "Lata szkolne",
    grade: "Klasy",
    role: "Role",
    course: "Kursy",
    coursePosts: "Wpisy kursów",
    postAttachments: "Załączniki do wpisów",
    courseAssignments: "Zadania kursów",
    assignmentSubmissions: "Przesłane zadania",
    courseStudents: "Uczniowie kursów",
    student: "Uczniowie",
    log: "Dziennik audytu",
}

const polishPresets: Record<string, {name: string; description: string}> = {
    administrator: {name: "Administrator", description: "Pełny dostęp do wszystkich dostępnych uprawnień."},
    "academic-manager": {name: "Opiekun dydaktyczny", description: "Zarządza latami szkolnymi i strukturą klas."},
    "staff-manager": {name: "Opiekun personelu", description: "Zarządza dostępem personelu i przypisanymi rolami."},
    teacher: {name: "Nauczyciel", description: "Wyświetla szkołę, lata szkolne, klasy i kursy."},
}

export function localizedPermissionLabel(locale: Locale, permission: string, englishLabel: string) {
    return locale === "pl" ? polishPermissionLabels[permission] ?? englishLabel : englishLabel
}

export function localizedPermissionDescription(locale: Locale, label: string, englishDescription: string) {
    return locale === "pl" ? `Pozwala tej roli na: ${label.toLocaleLowerCase("pl-PL")}.` : englishDescription
}

export function localizedPermissionCategory(locale: Locale, category: string, englishLabel: string) {
    return locale === "pl" ? polishCategories[category] ?? englishLabel : englishLabel
}

export function localizedPermissionPreset(locale: Locale, id: string, name: string, description: string) {
    return locale === "pl" ? polishPresets[id] ?? {name, description} : {name, description}
}
