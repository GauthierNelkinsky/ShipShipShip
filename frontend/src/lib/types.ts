export type EventStatus =
  | "Backlogs"
  | "Doing"
  | "Release"
  | "Upcoming"
  | "Archived";

export interface Tag {
  name: string;
  color: string;
}

export interface Event {
  id: number;
  title: string;
  tags: string; // JSON string of array
  media: string; // JSON string of array
  status: EventStatus;
  date: string;
  votes: number;
  content: string; // Markdown content
  order: number; // Order for sorting within status
  created_at: string;
  updated_at: string;
}

export interface CreateEventRequest {
  title: string;
  tags: string[];
  media: string[];
  status: EventStatus;
  date: string;
  content: string;
  order?: number;
}

export interface UpdateEventRequest {
  title?: string;
  tags?: string[];
  media?: string[];
  status?: EventStatus;
  date?: string;
  content?: string;
  order?: number;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  token: string;
}

export interface ApiError {
  error: string;
}

export interface VoteResponse {
  message: string;
  votes: number;
}

// Parsed versions for easier use in components
export interface ParsedEvent extends Omit<Event, "tags" | "media"> {
  tags: string[];
  media: string[];
}

// Reorder request interface
export interface ReorderEventRequest {
  event_id: number;
  new_order: number;
  status: string;
}

// Settings types
export interface ProjectSettings {
  title: string;
  logo_url: string;
  dark_logo_url: string;
  favicon_url: string;
  website_url: string;
  primary_color: string;
  created_at: string;
  updated_at: string;
}

export interface UpdateSettingsRequest {
  title?: string;
  logo_url?: string;
  dark_logo_url?: string;
  favicon_url?: string;
  website_url?: string;
  primary_color?: string;
}

export interface TagResponse {
  tags: Tag[];
}

export interface CreateTagRequest {
  name: string;
  color: string;
}
