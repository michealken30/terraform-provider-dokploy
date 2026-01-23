package client

import "time"

// Project represents a Dokploy project containing environments
type Project struct {
	ProjectID      string        `json:"projectId"`
	Name           string        `json:"name"`
	Description    string        `json:"description"`
	CreatedAt      time.Time     `json:"createdAt"`
	OrganizationID string        `json:"organizationId"`
	Env            string        `json:"env"`
	Environments   []Environment `json:"environments"`
}

// Environment represents an environment within a project
type Environment struct {
	EnvironmentID string        `json:"environmentId"`
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	CreatedAt     time.Time     `json:"createdAt"`
	Env           string        `json:"env"`
	ProjectID     string        `json:"projectId"`
	IsDefault     bool          `json:"isDefault"`
	Applications  []Application `json:"applications"`
	Compose       []Compose     `json:"compose"`
	Postgres      []Postgres    `json:"postgres"`
	MySQL         []MySQL       `json:"mysql"`
	MariaDB       []MariaDB     `json:"mariadb"`
	Mongo         []Mongo       `json:"mongo"`
	Redis         []Redis       `json:"redis"`
}

// Application represents a web application service
type Application struct {
	ApplicationID     string    `json:"applicationId"`
	Name              string    `json:"name"`
	AppName           string    `json:"appName"`
	Description       string    `json:"description"`
	Env               string    `json:"env"`
	CreatedAt         time.Time `json:"createdAt"`
	EnvironmentID     string    `json:"environmentId"`
	SourceType        string    `json:"sourceType"`
	Repository        string    `json:"repository"`
	Owner             string    `json:"owner"`
	Branch            string    `json:"branch"`
	BuildPath         string    `json:"buildPath"`
	AutoDeploy        bool      `json:"autoDeploy"`
	BuildType         string    `json:"buildType"`
	Dockerfile        string    `json:"dockerfile"`
	Replicas          int       `json:"replicas"`
	ApplicationStatus string    `json:"applicationStatus"`
	Command           *string   `json:"command"`
	Args              *string   `json:"args"`
	MemoryReservation *int64    `json:"memoryReservation"`
	MemoryLimit       *int64    `json:"memoryLimit"`
	CPUReservation    *int64    `json:"cpuReservation"`
	CPULimit          *int64    `json:"cpuLimit"`
	Title             *string   `json:"title"`
	Enabled           *bool     `json:"enabled"`
	Subtitle          *string   `json:"subtitle"`
	BuildArgs         string    `json:"buildArgs"`
	BuildSecrets      string    `json:"buildSecrets"`
	WatchPaths        []string  `json:"watchPaths"`
	Domains           []Domain  `json:"domains"`
	Ports             []Port    `json:"ports"`
	Mounts            []Mount   `json:"mounts"`
	ServerID          *string   `json:"serverId"`
	RegistryID        *string   `json:"registryId"`
}

// Compose represents a docker-compose service
type Compose struct {
	ComposeID         string    `json:"composeId"`
	Name              string    `json:"name"`
	AppName           string    `json:"appName"`
	Description       string    `json:"description"`
	Env               string    `json:"env"`
	CreatedAt         time.Time `json:"createdAt"`
	EnvironmentID     string    `json:"environmentId"`
	ComposeFile       string    `json:"composeFile"`
	ComposeType       string    `json:"composeType"`
	ComposeStatus     string    `json:"composeStatus"`
	SourceType        string    `json:"sourceType"`
	Repository        string    `json:"repository"`
	Owner             string    `json:"owner"`
	Branch            string    `json:"branch"`
	AutoDeploy        bool      `json:"autoDeploy"`
	CustomGitURL      *string   `json:"customGitUrl"`
	CustomGitBranch   *string   `json:"customGitBranch"`
	CustomGitSSHKeyID *string   `json:"customGitSSHKeyId"`
	Command           string    `json:"command"`
	Suffix            string    `json:"suffix"`
	RandomizeCompose  bool      `json:"randomizeCompose"`
	ServerID          *string   `json:"serverId"`
}

// Postgres represents a PostgreSQL database service
type Postgres struct {
	PostgresID        string    `json:"postgresId"`
	Name              string    `json:"name"`
	AppName           string    `json:"appName"`
	Description       string    `json:"description"`
	CreatedAt         time.Time `json:"createdAt"`
	EnvironmentID     string    `json:"environmentId"`
	DatabaseName      string    `json:"databaseName"`
	DatabaseUser      string    `json:"databaseUser"`
	DatabasePassword  string    `json:"databasePassword"`
	DockerImage       string    `json:"dockerImage"`
	Command           *string   `json:"command"`
	Env               string    `json:"env"`
	MemoryReservation *int64    `json:"memoryReservation"`
	MemoryLimit       *int64    `json:"memoryLimit"`
	CPUReservation    *int64    `json:"cpuReservation"`
	CPULimit          *int64    `json:"cpuLimit"`
	ExternalPort      *int      `json:"externalPort"`
	ApplicationStatus string    `json:"applicationStatus"`
	ServerID          *string   `json:"serverId"`
}

// MySQL represents a MySQL database service
type MySQL struct {
	MySQLID              string    `json:"mysqlId"`
	Name                 string    `json:"name"`
	AppName              string    `json:"appName"`
	Description          string    `json:"description"`
	CreatedAt            time.Time `json:"createdAt"`
	EnvironmentID        string    `json:"environmentId"`
	DatabaseName         string    `json:"databaseName"`
	DatabaseUser         string    `json:"databaseUser"`
	DatabasePassword     string    `json:"databasePassword"`
	DatabaseRootPassword string    `json:"databaseRootPassword"`
	DockerImage          string    `json:"dockerImage"`
	Command              *string   `json:"command"`
	Env                  string    `json:"env"`
	MemoryReservation    *int64    `json:"memoryReservation"`
	MemoryLimit          *int64    `json:"memoryLimit"`
	CPUReservation       *int64    `json:"cpuReservation"`
	CPULimit             *int64    `json:"cpuLimit"`
	ExternalPort         *int      `json:"externalPort"`
	ApplicationStatus    string    `json:"applicationStatus"`
	ServerID             *string   `json:"serverId"`
}

// MariaDB represents a MariaDB database service
type MariaDB struct {
	MariaDBID            string    `json:"mariadbId"`
	Name                 string    `json:"name"`
	AppName              string    `json:"appName"`
	Description          string    `json:"description"`
	CreatedAt            time.Time `json:"createdAt"`
	EnvironmentID        string    `json:"environmentId"`
	DatabaseName         string    `json:"databaseName"`
	DatabaseUser         string    `json:"databaseUser"`
	DatabasePassword     string    `json:"databasePassword"`
	DatabaseRootPassword string    `json:"databaseRootPassword"`
	DockerImage          string    `json:"dockerImage"`
	Command              *string   `json:"command"`
	Env                  string    `json:"env"`
	MemoryReservation    *int64    `json:"memoryReservation"`
	MemoryLimit          *int64    `json:"memoryLimit"`
	CPUReservation       *int64    `json:"cpuReservation"`
	CPULimit             *int64    `json:"cpuLimit"`
	ExternalPort         *int      `json:"externalPort"`
	ApplicationStatus    string    `json:"applicationStatus"`
	ServerID             *string   `json:"serverId"`
}

// Mongo represents a MongoDB database service
type Mongo struct {
	MongoID           string    `json:"mongoId"`
	Name              string    `json:"name"`
	AppName           string    `json:"appName"`
	Description       string    `json:"description"`
	CreatedAt         time.Time `json:"createdAt"`
	EnvironmentID     string    `json:"environmentId"`
	DatabaseUser      string    `json:"databaseUser"`
	DatabasePassword  string    `json:"databasePassword"`
	DockerImage       string    `json:"dockerImage"`
	Command           *string   `json:"command"`
	Env               string    `json:"env"`
	MemoryReservation *int64    `json:"memoryReservation"`
	MemoryLimit       *int64    `json:"memoryLimit"`
	CPUReservation    *int64    `json:"cpuReservation"`
	CPULimit          *int64    `json:"cpuLimit"`
	ExternalPort      *int      `json:"externalPort"`
	ApplicationStatus string    `json:"applicationStatus"`
	ServerID          *string   `json:"serverId"`
}

// Redis represents a Redis database service
type Redis struct {
	RedisID           string    `json:"redisId"`
	Name              string    `json:"name"`
	AppName           string    `json:"appName"`
	Description       string    `json:"description"`
	CreatedAt         time.Time `json:"createdAt"`
	EnvironmentID     string    `json:"environmentId"`
	DatabasePassword  string    `json:"databasePassword"`
	DockerImage       string    `json:"dockerImage"`
	Command           *string   `json:"command"`
	Env               string    `json:"env"`
	MemoryReservation *int64    `json:"memoryReservation"`
	MemoryLimit       *int64    `json:"memoryLimit"`
	CPUReservation    *int64    `json:"cpuReservation"`
	CPULimit          *int64    `json:"cpuLimit"`
	ExternalPort      *int      `json:"externalPort"`
	ApplicationStatus string    `json:"applicationStatus"`
	ServerID          *string   `json:"serverId"`
}

// Domain represents a domain configuration
type Domain struct {
	DomainID        string    `json:"domainId"`
	Host            string    `json:"host"`
	Port            *int      `json:"port"`
	HTTPS           bool      `json:"https"`
	CertificateType string    `json:"certificateType"`
	ApplicationID   *string   `json:"applicationId"`
	CreatedAt       time.Time `json:"createdAt"`
	Path            string    `json:"path"`
	UniqueConfigKey int       `json:"uniqueConfigKey"`
}

// Port represents a port mapping
type Port struct {
	PortID        string  `json:"portId"`
	PublishedPort int     `json:"publishedPort"`
	TargetPort    int     `json:"targetPort"`
	Protocol      string  `json:"protocol"`
	ApplicationID *string `json:"applicationId"`
}

// Mount represents a volume mount
type Mount struct {
	MountID       string  `json:"mountId"`
	Type          string  `json:"type"`
	HostPath      *string `json:"hostPath"`
	VolumeName    *string `json:"volumeName"`
	Content       *string `json:"content"`
	ServiceConfig any     `json:"serviceConfig"`
	MountPath     string  `json:"mountPath"`
	ApplicationID *string `json:"applicationId"`
}

// Server represents a server configuration
type Server struct {
	ServerID            string        `json:"serverId"`
	Name                string        `json:"name"`
	Description         string        `json:"description"`
	IPAddress           string        `json:"ipAddress"`
	Port                int           `json:"port"`
	Username            string        `json:"username"`
	AppName             string        `json:"appName"`
	EnableDockerCleanup bool          `json:"enableDockerCleanup"`
	CreatedAt           time.Time     `json:"createdAt"`
	OrganizationID      string        `json:"organizationId"`
	ServerStatus        string        `json:"serverStatus"`
	ServerType          string        `json:"serverType"`
	Command             string        `json:"command"`
	SSHKeyID            string        `json:"sshKeyId"`
	MetricsConfig       MetricsConfig `json:"metricsConfig"`
	TotalSum            int           `json:"totalSum"`
}

// MetricsConfig represents server metrics configuration
type MetricsConfig struct {
	Server     ServerMetrics    `json:"server"`
	Containers ContainerMetrics `json:"containers"`
}

// ServerMetrics represents server-level metrics config
type ServerMetrics struct {
	Port          int              `json:"port"`
	Type          string           `json:"type"`
	Token         string           `json:"token"`
	CronJob       string           `json:"cronJob"`
	Thresholds    MetricThresholds `json:"thresholds"`
	RefreshRate   int              `json:"refreshRate"`
	URLCallback   string           `json:"urlCallback"`
	RetentionDays int              `json:"retentionDays"`
}

// MetricThresholds represents threshold configuration
type MetricThresholds struct {
	CPU    int `json:"cpu"`
	Memory int `json:"memory"`
}

// ContainerMetrics represents container metrics config
type ContainerMetrics struct {
	Services    ServiceFilter `json:"services"`
	RefreshRate int           `json:"refreshRate"`
}

// ServiceFilter represents service filtering configuration
type ServiceFilter struct {
	Exclude []string `json:"exclude"`
	Include []string `json:"include"`
}

// SSHKey represents an SSH key
type SSHKey struct {
	SSHKeyID       string    `json:"sshKeyId"`
	PrivateKey     string    `json:"privateKey"`
	PublicKey      string    `json:"publicKey"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	CreatedAt      time.Time `json:"createdAt"`
	LastUsedAt     time.Time `json:"lastUsedAt"`
	OrganizationID string    `json:"organizationId"`
}

// Registry represents a container registry
type Registry struct {
	RegistryID     string    `json:"registryId"`
	RegistryName   string    `json:"registryName"`
	Username       string    `json:"username"`
	Password       string    `json:"password"`
	RegistryURL    string    `json:"registryUrl"`
	CreatedAt      time.Time `json:"createdAt"`
	SelfHosted     string    `json:"selfHosted"`
	ImagePrefix    *string   `json:"imagePrefix"`
	OrganizationID string    `json:"organizationId"`
	RegistryType   string    `json:"registryType"`
}

// Certificate represents an SSL certificate
type Certificate struct {
	CertificateID   string    `json:"certificateId"`
	Name            string    `json:"name"`
	CertificateData string    `json:"certificateData"`
	PrivateKey      string    `json:"privateKey"`
	CertificatePath string    `json:"certificatePath"`
	AutoRenew       bool      `json:"autoRenew"`
	CreatedAt       time.Time `json:"createdAt"`
	OrganizationID  string    `json:"organizationId"`
}

// Destination represents a backup destination (S3, etc.)
type Destination struct {
	DestinationID   string    `json:"destinationId"`
	Name            string    `json:"name"`
	AccessKey       string    `json:"accessKey"`
	SecretAccessKey string    `json:"secretAccessKey"`
	Bucket          string    `json:"bucket"`
	Region          string    `json:"region"`
	Endpoint        string    `json:"endpoint"`
	CreatedAt       time.Time `json:"createdAt"`
	OrganizationID  string    `json:"organizationId"`
}

// Notification represents a notification configuration
type Notification struct {
	NotificationID  string          `json:"notificationId"`
	Name            string          `json:"name"`
	Type            string          `json:"type"`
	AppDeploy       bool            `json:"appDeploy"`
	AppBuildError   bool            `json:"appBuildError"`
	DatabaseBackup  bool            `json:"databaseBackup"`
	DokployRestart  bool            `json:"dokployRestart"`
	DockerCleanup   bool            `json:"dockerCleanup"`
	ServerThreshold bool            `json:"serverThreshold"`
	CreatedAt       time.Time       `json:"createdAt"`
	OrganizationID  string          `json:"organizationId"`
	SlackConfig     *SlackConfig    `json:"slack,omitempty"`
	DiscordConfig   *DiscordConfig  `json:"discord,omitempty"`
	EmailConfig     *EmailConfig    `json:"email,omitempty"`
	TelegramConfig  *TelegramConfig `json:"telegram,omitempty"`
}

// SlackConfig represents Slack notification config
type SlackConfig struct {
	WebhookURL string `json:"webhookUrl"`
	Channel    string `json:"channel"`
}

// DiscordConfig represents Discord notification config
type DiscordConfig struct {
	WebhookURL string `json:"webhookUrl"`
}

// EmailConfig represents email notification config
type EmailConfig struct {
	SMTPServer string   `json:"smtpServer"`
	SMTPPort   int      `json:"smtpPort"`
	Username   string   `json:"username"`
	Password   string   `json:"password"`
	FromEmail  string   `json:"fromEmail"`
	ToEmails   []string `json:"toEmails"`
}

// TelegramConfig represents Telegram notification config
type TelegramConfig struct {
	BotToken string `json:"botToken"`
	ChatID   string `json:"chatId"`
}
