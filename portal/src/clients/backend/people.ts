import { axiosClient } from "./client";

interface Person {
    personId: string;
    name: string;
    age: string;
    imagePath: string;
}

/**
 * Get detectable people
 * @api GET /api/people
 * @param ids Person IDs
 * @returns People
 */
export async function getPeople(ids: string[]): Promise<Person[]> {
    let uri = "/api/people";
    if (ids.length > 0) {
        uri += '?ids=' + ids.join(',');
    }
    const resp = await axiosClient.get(uri);
    return resp.data["people"];
}

interface PeopleImage {
    presignedUrl: string;
    expires: string;
}

/**
 * Get people image
 * @api GET /api/people/presigned
 * @param imagePath Image path
 * @returns People image
 */
export async function getPeopleImage(personId: string): Promise<PeopleImage> {
    const uri = `/api/people/presigned?id=${personId}`;
    const resp = await axiosClient.get(uri);
    return resp.data;
}

export interface AddDetectablePerson {
    name: string;
    age: string;
    base64Image: string;
}

/**
 * Add person
 * @api POST /api/people
 * @param person Person information
 * @returns 
 */
export async function addDetectablePerson(person: AddDetectablePerson): Promise<void> {
    const uri = `/api/people`
    await axiosClient.post(uri, person);
    return;
}

/**
 * Delete person
 * @api DELETE /api/people
 * @param personId Person ID
 * @returns 
 */
export async function deletePerson(personId: string): Promise<void> {
    const uri = `/api/people?id=${personId}`
    await axiosClient.delete(uri);
    return
}

export interface PersonHistory {
    historyId: string;
    timestamp: string;
    eventId: string;
    personId: string;
}

/**
 * Get person history
 * @api GET /api/people/history
 * @param personIds List of person Ids
 * @returns List of PersonHistory
 */
export async function getPersonHistory(personIds: string[]): Promise<PersonHistory[]> {
    const uri = `/api/people/history?person_id=${personIds.join(",")}`
    const resp = await axiosClient.get(uri);
    return resp.data["histories"];
}