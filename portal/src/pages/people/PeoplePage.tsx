import createDebounce from "@solid-primitives/debounce";
import { Component, createResource, createSignal } from "solid-js";

export const PeoplePage: Component = () => {
    const [peopleIds, setPeopleIds] = createSignal<string[]>([]);

    const filterDebouncer = createDebounce((m: string) => {
        if (m === "") {
            setPeopleIds([]);
        } else {
            setPeopleIds(m.split(","));
        }
    }, 350);
    const [people, {refetch: fetchPeople}] = createResource(peopleIds, getList)
}