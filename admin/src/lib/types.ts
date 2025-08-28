export type EventStatus =
  | "Backlogs"
  | "Proposed"
  | "Upcoming"
  | "Release"
  | "Archived";

export interface Tag {
  id: number;
  name: string;
  color: string;
  created_at: string;
  updated_at: string;
}

export interface TagUsage {
  id: number;
  name: string;
  color: string;
  count: number;
}

export interface Event {
  id: number;
  title: string;
  tags: Tag[]; // Array of Tag objects
  media: string; // JSON string of array
  status: EventStatus;
  date: string;
  votes: number;
  content: string; // Markdown content
  order: number; // Order for sorting within status
  created_at: string;
  updated_at: string;
  is_public: boolean; // Controls if event appears on public page
  has_public_url: boolean; // Controls if event has individual public URL
  slug: string;
}

export interface CreateEventRequest {
  title: string;
  tag_ids: number[]; // Array of tag IDs instead of strings
  media: string[];
  status: EventStatus;
  date: string;
  content: string;
  order?: number;
}

export interface UpdateEventRequest {
  title?: string;
  tag_ids?: number[]; // Array of tag IDs instead of strings
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
export interface ParsedEvent extends Omit<Event, "media"> {
  media: string[];
  slug: string;
}

// Tag-related request types
export interface CreateTagRequest {
  name: string;
  color: string;
}

export interface UpdateTagRequest {
  name?: string;
  color?: string;
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
  newsletter_enabled: boolean;
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
  newsletter_enabled?: boolean;
}

// Mail settings types
export interface MailSettings {
  id: number;
  smtp_host: string;
  smtp_port: number;
  smtp_username: string;
  smtp_password: string;
  smtp_encryption: string;
  from_email: string;
  from_name: string;
  created_at: string;
  updated_at: string;
}

export interface UpdateMailSettingsRequest {
  smtp_host?: string;
  smtp_port?: number;
  smtp_username?: string;
  smtp_password?: string;
  smtp_encryption?: string;
  from_email?: string;
  from_name?: string;
}

// Footer Link types
export type FooterColumnType = "left" | "middle" | "right";

export interface FooterLink {
  id: number;
  name: string;
  url: string;
  column: FooterColumnType;
  order: number;
  open_in_new_window: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateFooterLinkRequest {
  name: string;
  url: string;
  column: FooterColumnType;
  open_in_new_window?: boolean;
}

export interface UpdateFooterLinkRequest {
  name?: string;
  url?: string;
  column?: FooterColumnType;
  open_in_new_window?: boolean;
}

export interface ReorderFooterLinksRequest {
  links: {
    id: number;
    order: number;
  }[];
}

// Newsletter automation settings types
export interface NewsletterAutomationSettings {
  id?: number;
  enabled: boolean;
  trigger_statuses: EventStatus[];
  created_at?: string;
  updated_at?: string;
}

export interface UpdateNewsletterAutomationRequest {
  enabled?: boolean;
  trigger_statuses?: EventStatus[];
}
