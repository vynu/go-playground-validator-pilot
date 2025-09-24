// Package models contains Slack webhook payload models with comprehensive validation rules.
// This module defines all Slack-specific data structures for webhook validation.
package models

// No imports needed for this model

// SlackMessagePayload represents the top-level Slack webhook payload structure.
// This handles various Slack webhook types including commands, events, and interactive components.
type SlackMessagePayload struct {
	Token       string       `json:"token" validate:"required,slack_token"`
	TeamID      string       `json:"team_id" validate:"required,alphanum"`
	TeamDomain  string       `json:"team_domain" validate:"required,hostname"`
	ChannelID   string       `json:"channel_id" validate:"required,alphanum"`
	ChannelName string       `json:"channel_name" validate:"required,slack_channel_name"`
	UserID      string       `json:"user_id" validate:"required,alphanum"`
	UserName    string       `json:"user_name" validate:"required,min=1,max=21,alphanum"`
	Command     string       `json:"command" validate:"omitempty,slack_command"`
	Text        string       `json:"text" validate:"omitempty,max=3000"`
	ResponseURL string       `json:"response_url" validate:"required,url"`
	TriggerID   string       `json:"trigger_id" validate:"omitempty"`
	Timestamp   string       `json:"ts" validate:"omitempty,slack_timestamp"`
	Message     SlackMessage `json:"message" validate:"omitempty"`
	Event       SlackEvent   `json:"event" validate:"omitempty"`
	Challenge   string       `json:"challenge" validate:"omitempty"` // For URL verification
	Type        string       `json:"type" validate:"required,oneof=url_verification event_callback command interactive_component"`
}

// SlackMessage represents a Slack message with comprehensive structure validation.
type SlackMessage struct {
	Type        string            `json:"type" validate:"required,oneof=message"`
	Subtype     string            `json:"subtype" validate:"omitempty"`
	Text        string            `json:"text" validate:"required,max=40000"`
	User        string            `json:"user" validate:"required,alphanum"`
	Timestamp   string            `json:"ts" validate:"required,slack_timestamp"`
	ThreadTS    string            `json:"thread_ts" validate:"omitempty,slack_timestamp"`
	Channel     string            `json:"channel" validate:"required,alphanum"`
	Attachments []SlackAttachment `json:"attachments" validate:"omitempty,dive,max=20"`
	Blocks      []SlackBlock      `json:"blocks" validate:"omitempty,dive,max=50"`
	Files       []SlackFile       `json:"files" validate:"omitempty,dive"`
	Reactions   []SlackReaction   `json:"reactions" validate:"omitempty,dive"`
	Replies     []SlackReply      `json:"replies" validate:"omitempty,dive"`
	IsStarred   bool              `json:"is_starred"`
	Pinned      bool              `json:"pinned_to"`
	Edited      *SlackEditInfo    `json:"edited" validate:"omitempty"`
}

// SlackEvent represents Slack event data for event subscriptions.
type SlackEvent struct {
	Type           string    `json:"type" validate:"required,oneof=message reaction_added reaction_removed channel_created channel_deleted user_joined user_left"`
	EventTimestamp string    `json:"event_ts" validate:"required,slack_timestamp"`
	User           string    `json:"user" validate:"omitempty,alphanum"`
	Channel        string    `json:"channel" validate:"omitempty,alphanum"`
	Item           SlackItem `json:"item" validate:"omitempty"`
	Reaction       string    `json:"reaction" validate:"omitempty,min=1,max=100"`
	Text           string    `json:"text" validate:"omitempty,max=40000"`
	Timestamp      string    `json:"ts" validate:"omitempty,slack_timestamp"`
}

// SlackAttachment represents rich message attachments with validation constraints.
type SlackAttachment struct {
	ID             int64         `json:"id" validate:"omitempty,gt=0"`
	Color          string        `json:"color" validate:"omitempty,oneof=good warning danger|hexcolor"`
	Fallback       string        `json:"fallback" validate:"omitempty,max=300"`
	AuthorName     string        `json:"author_name" validate:"omitempty,max=100"`
	AuthorLink     string        `json:"author_link" validate:"omitempty,url"`
	AuthorIcon     string        `json:"author_icon" validate:"omitempty,url"`
	Title          string        `json:"title" validate:"omitempty,max=300"`
	TitleLink      string        `json:"title_link" validate:"omitempty,url"`
	Text           string        `json:"text" validate:"omitempty,max=8000"`
	ImageURL       string        `json:"image_url" validate:"omitempty,url"`
	ThumbURL       string        `json:"thumb_url" validate:"omitempty,url"`
	Footer         string        `json:"footer" validate:"omitempty,max=300"`
	FooterIcon     string        `json:"footer_icon" validate:"omitempty,url"`
	Timestamp      int64         `json:"ts" validate:"omitempty,gt=0"`
	Fields         []SlackField  `json:"fields" validate:"omitempty,dive,max=10"`
	Actions        []SlackAction `json:"actions" validate:"omitempty,dive,max=5"`
	CallbackID     string        `json:"callback_id" validate:"omitempty,max=255"`
	AttachmentType string        `json:"attachment_type" validate:"omitempty,oneof=default"`
}

// SlackBlock represents Block Kit elements for rich message formatting.
type SlackBlock struct {
	Type      string              `json:"type" validate:"required,oneof=section divider image actions context header input file"`
	Text      *SlackBlockText     `json:"text" validate:"omitempty"`
	Fields    []SlackBlockText    `json:"fields" validate:"omitempty,dive,max=10"`
	Elements  []SlackBlockElement `json:"elements" validate:"omitempty,dive,max=10"`
	BlockID   string              `json:"block_id" validate:"omitempty,max=255"`
	Accessory *SlackBlockElement  `json:"accessory" validate:"omitempty"`
}

// SlackBlockText represents text objects in Block Kit.
type SlackBlockText struct {
	Type     string `json:"type" validate:"required,oneof=plain_text mrkdwn"`
	Text     string `json:"text" validate:"required,max=3000"`
	Emoji    bool   `json:"emoji"`
	Verbatim bool   `json:"verbatim"`
}

// SlackBlockElement represents interactive elements in Block Kit.
type SlackBlockElement struct {
	Type         string              `json:"type" validate:"required,oneof=button static_select external_select users_select conversations_select channels_select image plain_text_input datepicker timepicker radio_buttons checkboxes overflow"`
	Text         *SlackBlockText     `json:"text" validate:"omitempty"`
	Value        string              `json:"value" validate:"omitempty,max=2000"`
	URL          string              `json:"url" validate:"omitempty,url"`
	ActionID     string              `json:"action_id" validate:"omitempty,max=255"`
	Confirm      *SlackConfirmDialog `json:"confirm" validate:"omitempty"`
	Options      []SlackOption       `json:"options" validate:"omitempty,dive,max=100"`
	OptionGroups []SlackOptionGroup  `json:"option_groups" validate:"omitempty,dive,max=100"`
	InitialValue string              `json:"initial_value" validate:"omitempty,max=2000"`
	Multiline    bool                `json:"multiline"`
	MinLength    int                 `json:"min_length" validate:"omitempty,gte=0,lte=3000"`
	MaxLength    int                 `json:"max_length" validate:"omitempty,gte=1,lte=3000"`
	Placeholder  *SlackBlockText     `json:"placeholder" validate:"omitempty"`
	ImageURL     string              `json:"image_url" validate:"omitempty,url"`
	AltText      string              `json:"alt_text" validate:"omitempty,max=2000"`
}

// SlackField represents attachment fields with title-value pairs.
type SlackField struct {
	Title string `json:"title" validate:"required,max=300"`
	Value string `json:"value" validate:"required,max=2000"`
	Short bool   `json:"short"`
}

// SlackAction represents legacy interactive message actions.
type SlackAction struct {
	Name         string              `json:"name" validate:"required,max=300"`
	Text         string              `json:"text" validate:"required,max=300"`
	Type         string              `json:"type" validate:"required,oneof=button select"`
	Value        string              `json:"value" validate:"omitempty,max=2000"`
	Style        string              `json:"style" validate:"omitempty,oneof=default primary danger"`
	URL          string              `json:"url" validate:"omitempty,url"`
	Options      []SlackActionOption `json:"options" validate:"omitempty,dive,max=100"`
	OptionGroups []SlackOptionGroup  `json:"option_groups" validate:"omitempty,dive,max=100"`
	Confirm      *SlackConfirmDialog `json:"confirm" validate:"omitempty"`
	DataSource   string              `json:"data_source" validate:"omitempty,oneof=static users channels conversations external"`
}

// SlackActionOption represents options for select-type actions.
type SlackActionOption struct {
	Text        string `json:"text" validate:"required,max=300"`
	Value       string `json:"value" validate:"required,max=2000"`
	Description string `json:"description" validate:"omitempty,max=300"`
}

// SlackOption represents options for Block Kit select elements.
type SlackOption struct {
	Text        *SlackBlockText `json:"text" validate:"required"`
	Value       string          `json:"value" validate:"required,max=75"`
	Description *SlackBlockText `json:"description" validate:"omitempty"`
	URL         string          `json:"url" validate:"omitempty,url"`
}

// SlackOptionGroup represents option groups for select elements.
type SlackOptionGroup struct {
	Label   *SlackBlockText `json:"label" validate:"required"`
	Options []SlackOption   `json:"options" validate:"required,dive,max=100"`
}

// SlackConfirmDialog represents confirmation dialogs for dangerous actions.
type SlackConfirmDialog struct {
	Title   *SlackBlockText `json:"title" validate:"required"`
	Text    *SlackBlockText `json:"text" validate:"required"`
	Confirm *SlackBlockText `json:"confirm" validate:"required"`
	Deny    *SlackBlockText `json:"deny" validate:"required"`
	Style   string          `json:"style" validate:"omitempty,oneof=primary danger"`
}

// SlackFile represents file attachments with metadata and validation.
type SlackFile struct {
	ID                 string       `json:"id" validate:"required,alphanum"`
	Name               string       `json:"name" validate:"required,max=255"`
	Title              string       `json:"title" validate:"omitempty,max=255"`
	Mimetype           string       `json:"mimetype" validate:"required"`
	Filetype           string       `json:"filetype" validate:"required"`
	PrettyType         string       `json:"pretty_type" validate:"omitempty"`
	User               string       `json:"user" validate:"required,alphanum"`
	Mode               string       `json:"mode" validate:"omitempty,oneof=hosted external snippet post"`
	Editable           bool         `json:"editable"`
	IsExternal         bool         `json:"is_external"`
	ExternalType       string       `json:"external_type" validate:"omitempty"`
	Size               int64        `json:"size" validate:"gte=0"`
	URLPrivate         string       `json:"url_private" validate:"omitempty,url"`
	URLPrivateDownload string       `json:"url_private_download" validate:"omitempty,url"`
	Permalink          string       `json:"permalink" validate:"omitempty,url"`
	EditLink           string       `json:"edit_link" validate:"omitempty,url"`
	Preview            string       `json:"preview" validate:"omitempty"`
	PreviewHighlight   string       `json:"preview_highlight" validate:"omitempty"`
	Lines              int          `json:"lines" validate:"gte=0"`
	LinesMore          int          `json:"lines_more" validate:"gte=0"`
	IsPublic           bool         `json:"is_public"`
	PublicURLShared    bool         `json:"public_url_shared"`
	Channels           []string     `json:"channels" validate:"omitempty,dive,alphanum"`
	Groups             []string     `json:"groups" validate:"omitempty,dive,alphanum"`
	IMs                []string     `json:"ims" validate:"omitempty,dive,alphanum"`
	InitialComment     SlackComment `json:"initial_comment" validate:"omitempty"`
	CommentsCount      int          `json:"comments_count" validate:"gte=0"`
}

// SlackComment represents file comments with validation.
type SlackComment struct {
	ID        string `json:"id" validate:"required,alphanum"`
	Timestamp string `json:"timestamp" validate:"required,slack_timestamp"`
	User      string `json:"user" validate:"required,alphanum"`
	Comment   string `json:"comment" validate:"required,max=8000"`
}

// SlackReaction represents message reactions with user lists.
type SlackReaction struct {
	Name  string   `json:"name" validate:"required,max=100"`
	Count int      `json:"count" validate:"required,gt=0"`
	Users []string `json:"users" validate:"required,dive,alphanum"`
}

// SlackReply represents thread replies with user and timestamp.
type SlackReply struct {
	User      string `json:"user" validate:"required,alphanum"`
	Timestamp string `json:"ts" validate:"required,slack_timestamp"`
}

// SlackEditInfo represents message edit metadata.
type SlackEditInfo struct {
	User      string `json:"user" validate:"required,alphanum"`
	Timestamp string `json:"ts" validate:"required,slack_timestamp"`
}

// SlackItem represents items for reactions and other events.
type SlackItem struct {
	Type    string `json:"type" validate:"required,oneof=message file comment"`
	Channel string `json:"channel" validate:"required,alphanum"`
	TS      string `json:"ts" validate:"required,slack_timestamp"`
	File    string `json:"file" validate:"omitempty,alphanum"`
}

// SlackWorkspace represents workspace/team information.
type SlackWorkspace struct {
	ID     string `json:"id" validate:"required,alphanum"`
	Name   string `json:"name" validate:"required,min=1,max=255"`
	Domain string `json:"domain" validate:"required,hostname"`
	URL    string `json:"url" validate:"required,url"`
}

// SlackChannel represents channel information with metadata.
type SlackChannel struct {
	ID                 string       `json:"id" validate:"required,alphanum"`
	Name               string       `json:"name" validate:"required,slack_channel_name"`
	IsChannel          bool         `json:"is_channel"`
	IsGroup            bool         `json:"is_group"`
	IsIM               bool         `json:"is_im"`
	IsMpIM             bool         `json:"is_mpim"`
	IsPrivate          bool         `json:"is_private"`
	IsArchived         bool         `json:"is_archived"`
	IsGeneral          bool         `json:"is_general"`
	IsShared           bool         `json:"is_shared"`
	IsExtShared        bool         `json:"is_ext_shared"`
	IsOrgShared        bool         `json:"is_org_shared"`
	IsPendingExtShared bool         `json:"is_pending_ext_shared"`
	Creator            string       `json:"creator" validate:"omitempty,alphanum"`
	Created            int64        `json:"created" validate:"omitempty,gt=0"`
	Topic              SlackTopic   `json:"topic" validate:"omitempty"`
	Purpose            SlackPurpose `json:"purpose" validate:"omitempty"`
	NumMembers         int          `json:"num_members" validate:"gte=0"`
}

// SlackTopic represents channel topic information.
type SlackTopic struct {
	Value   string `json:"value" validate:"omitempty,max=250"`
	Creator string `json:"creator" validate:"omitempty,alphanum"`
	LastSet int64  `json:"last_set" validate:"omitempty,gt=0"`
}

// SlackPurpose represents channel purpose information.
type SlackPurpose struct {
	Value   string `json:"value" validate:"omitempty,max=250"`
	Creator string `json:"creator" validate:"omitempty,alphanum"`
	LastSet int64  `json:"last_set" validate:"omitempty,gt=0"`
}

// SlackUser represents Slack user profile information.
type SlackUser struct {
	ID                string       `json:"id" validate:"required,alphanum"`
	TeamID            string       `json:"team_id" validate:"required,alphanum"`
	Name              string       `json:"name" validate:"required,min=1,max=21"`
	RealName          string       `json:"real_name" validate:"omitempty,max=255"`
	DisplayName       string       `json:"display_name" validate:"omitempty,max=255"`
	Email             string       `json:"email" validate:"omitempty,email"`
	IsBot             bool         `json:"is_bot"`
	IsAdmin           bool         `json:"is_admin"`
	IsOwner           bool         `json:"is_owner"`
	IsPrimaryOwner    bool         `json:"is_primary_owner"`
	IsRestricted      bool         `json:"is_restricted"`
	IsUltraRestricted bool         `json:"is_ultra_restricted"`
	HasFiles          bool         `json:"has_files"`
	Presence          string       `json:"presence" validate:"omitempty,oneof=active away"`
	Profile           SlackProfile `json:"profile" validate:"omitempty"`
	Updated           int64        `json:"updated" validate:"omitempty,gt=0"`
}

// SlackProfile represents detailed user profile information.
type SlackProfile struct {
	Title                 string `json:"title" validate:"omitempty,max=255"`
	Phone                 string `json:"phone" validate:"omitempty,max=50"`
	Skype                 string `json:"skype" validate:"omitempty,max=255"`
	RealName              string `json:"real_name" validate:"omitempty,max=255"`
	RealNameNormalized    string `json:"real_name_normalized" validate:"omitempty,max=255"`
	DisplayName           string `json:"display_name" validate:"omitempty,max=255"`
	DisplayNameNormalized string `json:"display_name_normalized" validate:"omitempty,max=255"`
	StatusText            string `json:"status_text" validate:"omitempty,max=100"`
	StatusEmoji           string `json:"status_emoji" validate:"omitempty,max=100"`
	StatusExpiration      int64  `json:"status_expiration" validate:"omitempty,gte=0"`
	AvatarHash            string `json:"avatar_hash" validate:"omitempty"`
	Email                 string `json:"email" validate:"omitempty,email"`
	Image24               string `json:"image_24" validate:"omitempty,url"`
	Image32               string `json:"image_32" validate:"omitempty,url"`
	Image48               string `json:"image_48" validate:"omitempty,url"`
	Image72               string `json:"image_72" validate:"omitempty,url"`
	Image192              string `json:"image_192" validate:"omitempty,url"`
	Image512              string `json:"image_512" validate:"omitempty,url"`
	ImageOriginal         string `json:"image_original" validate:"omitempty,url"`
	FirstName             string `json:"first_name" validate:"omitempty,max=255"`
	LastName              string `json:"last_name" validate:"omitempty,max=255"`
}
