import type {
  Event,
  CreateEventRequest,
  UpdateEventRequest,
  ProjectSettings,
  UpdateSettingsRequest,
  MailSettings,
  UpdateMailSettingsRequest,
  Tag,
  TagUsage,
  CreateTagRequest,
  UpdateTagRequest,
  FooterLink,
  CreateFooterLinkRequest,
  UpdateFooterLinkRequest,
  ReorderFooterLinksRequest,
  NewsletterAutomationSettings,
  UpdateNewsletterAutomationRequest,
} from "./types";

const API_BASE = "/api";

class ApiClient {
  private token: string | null = null;

  constructor() {
    // Load token from localStorage on initialization
    if (typeof window !== "undefined") {
      this.token = localStorage.getItem("auth_token");
    }
  }

  setToken(token: string) {
    this.token = token;
    if (typeof window !== "undefined") {
      localStorage.setItem("auth_token", token);
    }
  }

  clearToken() {
    this.token = null;
    if (typeof window !== "undefined") {
      localStorage.removeItem("auth_token");
    }
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {},
  ): Promise<T> {
    const url = `${API_BASE}${endpoint}`;

    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      ...((options.headers as Record<string, string>) || {}),
    };

    if (this.token) {
      headers.Authorization = `Bearer ${this.token}`;
    }

    const config: RequestInit = {
      ...options,
      headers,
    };

    try {
      const response = await fetch(url, config);

      if (!response.ok) {
        const errorData = await response
          .json()
          .catch(() => ({ error: "Network error" }));
        throw new Error(errorData.error || `HTTP ${response.status}`);
      }

      return await response.json();
    } catch (error) {
      console.error("API request failed:", error);
      throw error;
    }
  }

  // Auth endpoints
  async login(username: string, password: string) {
    const response = await this.request<{ token: string }>("/auth/login", {
      method: "POST",
      body: JSON.stringify({ username, password }),
    });

    this.setToken(response.token);
    return response;
  }

  async validateToken() {
    return this.request<{ valid: boolean; username: string }>(
      "/admin/validate",
    );
  }

  async checkDemoMode() {
    return this.request<{ demo_mode: boolean }>("/auth/demo-mode");
  }

  logout() {
    this.clearToken();
  }

  // Public event endpoints
  async getEvents() {
    return this.request<Event[]>("/events");
  }

  async getEvent(id: number) {
    return this.request<Event>(`/events/${id}`);
  }

  async getEventBySlug(slug: string) {
    return this.request<Event>(`/events/slug/${slug}`);
  }

  async voteEvent(id: number) {
    return this.request<{ message: string; votes: number; voted: boolean }>(
      `/events/${id}/vote`,
      {
        method: "POST",
      },
    );
  }

  async checkVoteStatus(id: number) {
    return this.request<{ voted: boolean; votes: number }>(
      `/events/${id}/vote-status`,
    );
  }

  async submitFeedback(title: string, content: string, formStartTime: number) {
    return this.request<{ message: string; id: number }>("/feedback", {
      method: "POST",
      body: JSON.stringify({ title, content, form_start_time: formStartTime }),
    });
  }

  // Admin event endpoints
  async getAllEvents() {
    return this.request<Event[]>("/admin/events");
  }

  async createEvent(event: CreateEventRequest) {
    return this.request<Event>("/admin/events", {
      method: "POST",
      body: JSON.stringify(event),
    });
  }

  async updateEvent(id: number, event: UpdateEventRequest) {
    return this.request<Event>(`/admin/events/${id}`, {
      method: "PUT",
      body: JSON.stringify(event),
    });
  }

  async deleteEvent(id: number) {
    return this.request<{ message: string }>(`/admin/events/${id}`, {
      method: "DELETE",
    });
  }

  async reorderEvent(eventId: number, newOrder: number, status: string) {
    return this.request<{ message: string }>("/admin/events/reorder", {
      method: "POST",
      body: JSON.stringify({ event_id: eventId, new_order: newOrder, status }),
    });
  }

  // Settings endpoints
  async getSettings() {
    return this.request<ProjectSettings>("/settings");
  }

  async updateSettings(settings: UpdateSettingsRequest) {
    return this.request<ProjectSettings>("/admin/settings", {
      method: "PUT",
      body: JSON.stringify(settings),
    });
  }

  // Upload endpoints
  async uploadImage(
    file: File,
  ): Promise<{ url: string; filename: string; size: number }> {
    const formData = new FormData();
    formData.append("image", file);

    const url = `${API_BASE}/admin/upload/image`;

    const headers: Record<string, string> = {};
    if (this.token) {
      headers.Authorization = `Bearer ${this.token}`;
    }

    const response = await fetch(url, {
      method: "POST",
      headers,
      body: formData,
    });

    if (!response.ok) {
      const errorData = await response
        .json()
        .catch(() => ({ error: "Upload failed" }));
      throw new Error(errorData.error || `HTTP ${response.status}`);
    }

    return await response.json();
  }

  // Mail settings endpoints
  async getMailSettings() {
    return this.request<MailSettings>("/admin/settings/mail");
  }

  async updateMailSettings(settings: UpdateMailSettingsRequest) {
    return this.request<MailSettings>("/admin/settings/mail", {
      method: "POST",
      body: JSON.stringify(settings),
    });
  }

  async testMailSettings(email: string) {
    return this.request<{ message: string }>("/admin/settings/mail/test", {
      method: "POST",
      body: JSON.stringify({ email }),
    });
  }

  // Newsletter endpoints
  async subscribeToNewsletter(email: string) {
    return this.request<{ message: string; email: string }>(
      "/newsletter/subscribe",
      {
        method: "POST",
        body: JSON.stringify({ email }),
      },
    );
  }

  async unsubscribeFromNewsletter(email: string) {
    return this.request<{ message: string }>("/newsletter/unsubscribe", {
      method: "POST",
      body: JSON.stringify({ email }),
    });
  }

  async checkNewsletterSubscription(email: string) {
    return this.request<{ subscribed: boolean; active: boolean }>(
      `/newsletter/status?email=${encodeURIComponent(email)}`,
    );
  }

  async getNewsletterStats() {
    return this.request<{ active_subscribers: number }>(
      "/admin/newsletter/stats",
    );
  }

  async getNewsletterSubscribers() {
    return this.request<{ subscribers: any[]; total: number }>(
      "/admin/newsletter/subscribers",
    );
  }

  async getNewsletterSubscribersPaginated(
    page: number = 1,
    limit: number = 10,
  ) {
    return this.request<{
      subscribers: any[];
      total: number;
      page: number;
      limit: number;
      total_pages: number;
    }>(`/admin/newsletter/subscribers/paginated?page=${page}&limit=${limit}`);
  }

  async getNewsletterHistory(page: number = 1, limit: number = 10) {
    return this.request<{
      newsletters: any[];
      total: number;
      page: number;
      limit: number;
      total_pages: number;
    }>(`/admin/newsletter/history?page=${page}&limit=${limit}`);
  }

  async getEmailTemplates() {
    return this.request<{
      templates: {
        [key: string]: {
          subject: string;
          content: string;
        };
      };
    }>("/admin/newsletter/templates");
  }

  async updateEmailTemplates(templates: {
    [key: string]: {
      subject: string;
      content: string;
    };
  }) {
    return this.request<{ message: string }>("/admin/newsletter/templates", {
      method: "PUT",
      body: JSON.stringify({ templates }),
    });
  }

  async deleteNewsletterSubscriber(email: string) {
    return this.request<{ message: string }>(
      `/admin/newsletter/subscribers/${encodeURIComponent(email)}`,
      {
        method: "DELETE",
      },
    );
  }

  // Event publishing endpoints
  async getEventPublishStatus(eventId: number) {
    return this.request<{
      is_public: boolean;
      has_public_url: boolean;
      email_sent: boolean;
      email_sent_at?: string;
      email_subject?: string;
      email_template?: string;
      subscriber_count?: number;
    }>(`/admin/events/${eventId}/publish`);
  }

  async updateEventPublicStatus(
    eventId: number,
    data: { is_public?: boolean; has_public_url?: boolean },
  ) {
    return this.request<{ message: string; updates: any }>(
      `/admin/events/${eventId}/publish`,
      {
        method: "PUT",
        body: JSON.stringify(data),
      },
    );
  }

  async getEventNewsletterPreview(eventId: number, template: string) {
    return this.request<{ subject: string; content: string }>(
      `/admin/events/${eventId}/newsletter/preview?template=${template}`,
    );
  }

  async getEventEmailHistory(eventId: number) {
    return this.request<{
      history: Array<{
        id: number;
        event_id: number;
        event_status: string;
        email_subject: string;
        email_template: string;
        subscriber_count: number;
        sent_at: string;
        created_at: string;
      }>;
    }>(`/admin/events/${eventId}/newsletter/history`);
  }

  async sendEventNewsletter(
    eventId: number,
    data: { subject: string; content: string; template: string },
  ) {
    return this.request<{
      message: string;
      subscribers_sent: number;
      total_subscribers: number;
    }>(`/admin/events/${eventId}/newsletter/send`, {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  // Tag endpoints
  async getTags() {
    return this.request<Tag[]>("/tags");
  }

  async getTag(id: number) {
    return this.request<Tag>(`/admin/tags/${id}`);
  }

  async getTagUsage() {
    return this.request<TagUsage[]>("/admin/tags/usage");
  }

  async createTag(tag: CreateTagRequest) {
    return this.request<Tag>("/admin/tags", {
      method: "POST",
      body: JSON.stringify(tag),
    });
  }

  async updateTag(id: number, tag: UpdateTagRequest) {
    return this.request<Tag>(`/admin/tags/${id}`, {
      method: "PUT",
      body: JSON.stringify(tag),
    });
  }

  async deleteTag(id: number) {
    return this.request<{ message: string }>(`/admin/tags/${id}`, {
      method: "DELETE",
    });
  }

  // Footer links endpoints
  async getFooterLinks() {
    return this.request<{ links: FooterLink[] }>("/admin/footer-links");
  }

  async getFooterLinksByColumn() {
    return this.request<{ links: { [key: string]: FooterLink[] } }>(
      "/footer-links/by-column",
    );
  }

  async getFooterLink(id: number) {
    return this.request<FooterLink>(`/admin/footer-links/${id}`);
  }

  async createFooterLink(footerLink: CreateFooterLinkRequest) {
    return this.request<FooterLink>("/admin/footer-links", {
      method: "POST",
      body: JSON.stringify(footerLink),
    });
  }

  async updateFooterLink(id: number, footerLink: UpdateFooterLinkRequest) {
    return this.request<FooterLink>(`/admin/footer-links/${id}`, {
      method: "PUT",
      body: JSON.stringify(footerLink),
    });
  }

  async deleteFooterLink(id: number) {
    return this.request<{ message: string }>(`/admin/footer-links/${id}`, {
      method: "DELETE",
    });
  }

  async reorderFooterLinks(reorderData: ReorderFooterLinksRequest) {
    return this.request<{ message: string }>("/admin/footer-links/reorder", {
      method: "POST",
      body: JSON.stringify(reorderData),
    });
  }

  // Newsletter automation endpoints
  async getNewsletterAutomationSettings() {
    return this.request<NewsletterAutomationSettings>(
      "/admin/newsletter/automation",
    );
  }

  async updateNewsletterAutomationSettings(
    settings: UpdateNewsletterAutomationRequest,
  ) {
    return this.request<NewsletterAutomationSettings>(
      "/admin/newsletter/automation",
      {
        method: "PUT",
        body: JSON.stringify(settings),
      },
    );
  }

  async fetchCurrentTheme() {
    try {
      const response = await fetch("/api/admin/themes/current", {
        headers: {
          Authorization: `Bearer ${this.token}`,
          "Content-Type": "application/json",
        },
      });
      if (response.ok) {
        const data = await response.json();
        return data; // Assuming data contains the current theme info
      }
    } catch (err) {
      console.error("Error fetching current theme:", err);
      // Don't fail the whole operation if we can't get current theme
    }
  }

  // Helper method to check if user is authenticated
  isAuthenticated(): boolean {
    return !!this.token;
  }
}

// Export a singleton instance
export const api = new ApiClient();

// Export types for convenience
export type {
  Event,
  CreateEventRequest,
  UpdateEventRequest,
  EventStatus,
  ParsedEvent,
  ProjectSettings,
  UpdateSettingsRequest,
  ReorderEventRequest,
  MailSettings,
  UpdateMailSettingsRequest,
  Tag,
  TagUsage,
  CreateTagRequest,
  UpdateTagRequest,
  FooterLink,
  CreateFooterLinkRequest,
  UpdateFooterLinkRequest,
  ReorderFooterLinksRequest,
  NewsletterAutomationSettings,
  UpdateNewsletterAutomationRequest,
} from "./types";
