import { error } from "@sveltejs/kit";
import { api } from "$lib/api";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ params }) => {
  const slug = params.event_slug;

  // Validate slug is not empty
  if (!slug || slug.trim() === "") {
    throw error(404, "Event not found");
  }

  try {
    const event = await api.getEventBySlug(slug);
    const settings = await api.getSettings();

    return {
      event,
      settings,
    };
  } catch (err) {
    console.error("Failed to load event:", err);
    throw error(404, "Event not found");
  }
};
