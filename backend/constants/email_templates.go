package constants

// EmailTemplateTypes defines the available email template types
const (
	TemplateTypeUpcomingFeature = "upcoming_feature"
	TemplateTypeNewRelease      = "new_release"
	TemplateTypeProposedFeature = "proposed_feature"
	TemplateTypeWelcome         = "welcome"
)

// Email template subjects
const (
	SubjectUpcomingFeature = "Coming Soon: {{event_name}} - {{project_name}}"
	SubjectNewRelease      = "🎉 New Release: {{event_name}} - {{project_name}}"
	SubjectProposedFeature = "💡 New Proposal: {{event_name}} - {{project_name}}"
	SubjectWelcome         = "Welcome to {{project_name}}!"
)

// Email template content
const (
	TemplateUpcomingFeature = `<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <h1 style="color: #000000; text-align: center; font-size: 28px; font-weight: bold; margin: 20px 0;">🚀 Coming Soon!</h1>

    <div style="padding: 20px; margin-bottom: 20px;">
        <h2 style="color: #000000; margin-top: 0; font-size: 22px; font-weight: bold; margin-bottom: 15px;">{{event_name}}</h2>

        <div style="margin-bottom: 20px;">
            <div style="margin-bottom: 8px; color: #6b7280; font-size: 14px;">
                {{event_date}}
            </div>
            <div style="display: flex; flex-wrap: wrap; gap: 6px; align-items: center;">
                {{event_tags}}
            </div>
        </div>

        <div style="margin: 15px 0; font-size: 16px; line-height: 1.6;">
            {{event_content}}
        </div>
        <div style="text-align: center; margin-top: 30px;">
            <a href="{{event_url}}" style="background: {{primary_color}}; color: white; padding: 14px 28px; text-decoration: none; border-radius: 6px; display: inline-block; font-weight: bold; font-size: 16px;">See Details</a>
        </div>
    </div>

    <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">

    <div style="text-align: center; font-size: 12px; color: #666;">
        <p style="margin: 5px 0;">
            <a href="{{project_url}}" style="color: #2563eb; text-decoration: none;">{{project_name}}</a>
            <br><a href="{{unsubscribe_url}}" style="color: #2563eb; text-decoration: none;">Unsubscribe</a>
        </p>
    </div>
</body>`

	TemplateNewRelease = `<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <h1 style="color: #000000; text-align: center; font-size: 28px; font-weight: bold; margin: 20px 0;">🎉 New Release Available!</h1>

    <div style="padding: 20px; margin-bottom: 20px;">
        <h2 style="color: #000000; margin-top: 0; font-size: 22px; font-weight: bold; margin-bottom: 15px;">{{event_name}}</h2>

        <div style="margin-bottom: 20px;">
            <div style="margin-bottom: 8px; color: #6b7280; font-size: 14px;">
                {{event_date}}
            </div>
            <div style="display: flex; flex-wrap: wrap; gap: 6px; align-items: center;">
                {{event_tags}}
            </div>
        </div>

        <div style="margin: 15px 0; font-size: 16px; line-height: 1.6;">
            {{event_content}}
        </div>
        <div style="text-align: center; margin-top: 30px;">
            <a href="{{event_url}}" style="background: {{primary_color}}; color: white; padding: 14px 28px; text-decoration: none; border-radius: 6px; display: inline-block; font-weight: bold; font-size: 16px;">See Details</a>
        </div>
    </div>

    <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">

    <div style="text-align: center; font-size: 12px; color: #666;">
        <p style="margin: 5px 0;">
            <a href="{{project_url}}" style="color: #2563eb; text-decoration: none;">{{project_name}}</a>
            <br><a href="{{unsubscribe_url}}" style="color: #2563eb; text-decoration: none;">Unsubscribe</a>
        </p>
    </div>
</body>`

	TemplateProposedFeature = `<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <h1 style="color: #000000; text-align: center; font-size: 28px; font-weight: bold; margin: 20px 0;">💡 New Feature Proposal!</h1>

    <div style="padding: 20px; margin-bottom: 20px;">
        <h2 style="color: #000000; margin-top: 0; font-size: 22px; font-weight: bold; margin-bottom: 15px;">{{event_name}}</h2>

        <div style="margin-bottom: 20px;">
            <div style="margin-bottom: 8px; color: #6b7280; font-size: 14px;">
                {{event_date}}
            </div>
            <div style="display: flex; flex-wrap: wrap; gap: 6px; align-items: center;">
                {{event_tags}}
            </div>
        </div>

        <div style="margin: 15px 0; font-size: 16px; line-height: 1.6;">
            {{event_content}}
        </div>
        <div style="text-align: center; margin-top: 30px;">
            <a href="{{event_url}}" style="background: {{primary_color}}; color: white; padding: 14px 28px; text-decoration: none; border-radius: 6px; display: inline-block; font-weight: bold; font-size: 16px;">Vote & See Details</a>
        </div>
    </div>

    <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">

    <div style="text-align: center; font-size: 12px; color: #666;">
        <p style="margin: 5px 0;">
            <a href="{{project_url}}" style="color: #2563eb; text-decoration: none;">{{project_name}}</a>
            <br><a href="{{unsubscribe_url}}" style="color: #2563eb; text-decoration: none;">Unsubscribe</a>
        </p>
    </div>
</body>`

	TemplateWelcome = `<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <h1 style="color: #000000; text-align: center; font-size: 28px; font-weight: bold; margin: 20px 0;">🎉 Welcome to {{project_name}}!</h1>

    <div style="padding: 20px; margin-bottom: 20px;">
        <h2 style="color: #000000; margin-top: 0; font-size: 22px; font-weight: bold; margin-bottom: 15px;">Thanks for subscribing!</h2>

        <div style="margin: 15px 0; font-size: 16px; line-height: 1.6;">
            You've successfully subscribed to our newsletter. We'll keep you updated with the latest features, releases, and important announcements.
        </div>
    </div>

    <hr style="border: none; border-top: 1px solid #eee; margin: 30px 0;">

    <div style="text-align: center; font-size: 12px; color: #666;">
        <p style="margin: 5px 0;">
            <a href="{{project_url}}" style="color: #2563eb; text-decoration: none;">{{project_name}}</a>
            <br><a href="{{unsubscribe_url}}" style="color: #2563eb; text-decoration: none;">Unsubscribe</a>
        </p>
    </div>
</body>`
)

// EmailTemplateData represents the structure for email template data
type EmailTemplateData struct {
	Type    string
	Subject string
	Content string
}

// GetDefaultTemplates returns all default email templates
func GetDefaultTemplates() []EmailTemplateData {
	return []EmailTemplateData{
		{
			Type:    TemplateTypeUpcomingFeature,
			Subject: SubjectUpcomingFeature,
			Content: TemplateUpcomingFeature,
		},
		{
			Type:    TemplateTypeNewRelease,
			Subject: SubjectNewRelease,
			Content: TemplateNewRelease,
		},
		{
			Type:    TemplateTypeProposedFeature,
			Subject: SubjectProposedFeature,
			Content: TemplateProposedFeature,
		},
		{
			Type:    TemplateTypeWelcome,
			Subject: SubjectWelcome,
			Content: TemplateWelcome,
		},
	}
}

// GetTemplateByType returns a specific template by type
func GetTemplateByType(templateType string) *EmailTemplateData {
	templates := GetDefaultTemplates()
	for _, template := range templates {
		if template.Type == templateType {
			return &template
		}
	}
	return nil
}
