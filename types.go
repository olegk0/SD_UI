package main

type ModelInfo struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Stem string `json:"stem"`
}

type LoraInfo struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type UpscalerInfo struct {
	Name string `json:"name"`
}

type LimitsInfo struct {
	MaxBatchCount int `json:"max_batch_count"`
	MaxHeight     int `json:"max_height"`
	MaxQueueSize  int `json:"max_queue_size"`
	MaxWidth      int `json:"max_width"`
	MinHeight     int `json:"min_height"`
	MinWidth      int `json:"min_width"`
}

type OutputFormatsByModeInfo struct {
	// Ключ карты — режим (например, "img_gen"), значение — список форматов
	OutputFormatsByMode map[string][]string `json:"output_formats_by_mode"`
}

type FeaturesByModeInfo struct {
	// Внешний ключ — режим ("img_gen"), внутренний ключ — фича ("cache"), значение — true/false
	FeaturesByMode map[string]map[string]bool `json:"features_by_mode"`
}

type DefaultsByModeInfo struct {
	DefaultsByMode map[string]ImgGenDefaults `json:"defaults_by_mode"`
}

// ImgGenDefaults описывает дефолтные настройки для генерации
type ImgGenDefaults struct {
	AutoResizeRefImage bool            `json:"auto_resize_ref_image"`
	BatchCount         int             `json:"batch_count"`
	CacheMode          string          `json:"cache_mode"`
	CacheOption        string          `json:"cache_option"`
	ClipSkip           int             `json:"clip_skip"`
	ControlStrength    float64         `json:"control_strength"`
	Height             int             `json:"height"`
	Hires              HiresConfig     `json:"hires"`
	IncreaseRefIndex   bool            `json:"increase_ref_index"`
	NegativePrompt     string          `json:"negative_prompt"`
	OutputCompression  int             `json:"output_compression"`
	OutputFormat       string          `json:"output_format"`
	Prompt             string          `json:"prompt"`
	SampleParams       SampleParams    `json:"sample_params"`
	ScmMask            string          `json:"scm_mask"`
	ScmPolicyDynamic   bool            `json:"scm_policy_dynamic"`
	Seed               int64           `json:"seed"`
	Strength           float64         `json:"strength"`
	VaeTilingParams    VaeTilingParams `json:"vae_tiling_params"`
	Width              int             `json:"width"`
}

type HiresConfig struct {
	CustomSigmas      []float64 `json:"custom_sigmas"` // Пустой массив в JSON
	DenoisingStrength float64   `json:"denoising_strength"`
	Enabled           bool      `json:"enabled"`
	Scale             float64   `json:"scale"`
	Steps             int       `json:"steps"`
	TargetHeight      int       `json:"target_height"`
	TargetWidth       int       `json:"target_width"`
	UpscaleTileSize   int       `json:"upscale_tile_size"`
	Upscaler          string    `json:"upscaler"`
}

type SampleParams struct {
	Eta             *float64       `json:"eta"`
	FlowShift       *float64       `json:"flow_shift"`
	Guidance        GuidanceConfig `json:"guidance"`
	SampleMethod    string         `json:"sample_method"`
	SampleSteps     int            `json:"sample_steps"`
	Scheduler       string         `json:"scheduler"`
	ShiftedTimestep int            `json:"shifted_timestep"`
}

type GuidanceConfig struct {
	DistilledGuidance float64   `json:"distilled_guidance"`
	ImgCfg            float64   `json:"img_cfg"`
	Slg               SlgConfig `json:"slg"`
	TxtCfg            float64   `json:"txt_cfg"`
}

type SlgConfig struct {
	LayerEnd   float64 `json:"layer_end"`
	LayerStart float64 `json:"layer_start"`
	Layers     []int   `json:"layers"`
	Scale      float64 `json:"scale"`
}

type VaeTilingParams struct {
	Enabled         bool    `json:"enabled"`
	ExtraTilingArgs string  `json:"extra_tiling_args"`
	RelSizeX        float64 `json:"rel_size_x"`
	RelSizeY        float64 `json:"rel_size_y"`
	TargetOverlap   float64 `json:"target_overlap"`
	TemporalTiling  bool    `json:"temporal_tiling"`
	TileSizeX       int     `json:"tile_size_x"`
	TileSizeY       int     `json:"tile_size_y"`
}

type FeaturesFlags struct {
	Cache            bool `json:"cache"`
	CancelGenerating bool `json:"cancel_generating"`
	CancelQueued     bool `json:"cancel_queued"`
	ControlImage     bool `json:"control_image"`
	Hires            bool `json:"hires"`
	InitImage        bool `json:"init_image"`
	Lora             bool `json:"lora"`
	MaskImage        bool `json:"mask_image"`
	RefImages        bool `json:"ref_images"`
	VaeTiling        bool `json:"vae_tiling"`
}

type CapabilitiesResponse struct {
	Model               ModelInfo               `json:"model"`
	CurrentMode         string                  `json:"current_mode"`
	SupportedModes      []string                `json:"supported_modes"`
	Defaults            ImgGenDefaults          `json:"defaults"`
	OutputFormats       []string                `json:"output_formats"`
	Features            FeaturesFlags           `json:"features"`
	DefaultsByMode      DefaultsByModeInfo      `json:"defaults_by_mode"`
	OutputFormatsByMode OutputFormatsByModeInfo `json:"output_formats_by_mode"`
	FeaturesByMode      FeaturesByModeInfo      `json:"features_by_mode"`
	Samplers            []string                `json:"samplers"`
	Schedulers          []string                `json:"schedulers"`
	Loras               []LoraInfo              `json:"loras"`
	Upscalers           []UpscalerInfo          `json:"upscalers"`
	Limits              LimitsInfo              `json:"limits"`
}

// ====================================== Запросы ===============================
type LoraParams struct {
	Name        string  `json:"name"`
	Path        string  `json:"path"`
	Multiplier  float64 `json:"multiplier"`
	IsHighNoise bool    `json:"is_high_noise"`
}

// -----------------------------------VidGenRequest-----------------------------------
type VidGenRequest struct {
	Prompt                string          `json:"prompt"`
	NegativePrompt        string          `json:"negative_prompt"`
	ClipSkip              int             `json:"clip_skip"`
	Width                 int             `json:"width"`
	Height                int             `json:"height"`
	Strength              float64         `json:"strength"`
	Seed                  int64           `json:"seed"`
	VideoFrames           int             `json:"video_frames"`
	Fps                   int             `json:"fps"`
	MoeBoundary           float64         `json:"moe_boundary"`
	VaceStrength          float64         `json:"vace_strength"`
	InitImage             *string         `json:"init_image"` // указатель для null
	EndImage              *string         `json:"end_image"`  // указатель для null
	ControlFrames         []string        `json:"control_frames"`
	SampleParams          SampleParams    `json:"sample_params"`
	HighNoiseSampleParams SampleParams    `json:"high_noise_sample_params"`
	Lora                  []LoraParams    `json:"lora"`
	VaeTilingParams       VaeTilingParams `json:"vae_tiling_params"`
	CacheMode             string          `json:"cache_mode"`
	CacheOption           string          `json:"cache_option"`
	ScmMask               string          `json:"scm_mask"`
	ScmPolicyDynamic      bool            `json:"scm_policy_dynamic"`
	OutputFormat          string          `json:"output_format"`
	OutputCompression     int             `json:"output_compression"`
}

// -----------------------------------ImgGenRequest-----------------------------------
type ImgGenRequest struct {
	Prompt             string  `json:"prompt"`
	NegativePrompt     string  `json:"negative_prompt"`
	ClipSkip           int     `json:"clip_skip"`
	Width              int     `json:"width"`
	Height             int     `json:"height"`
	Strength           float64 `json:"strength"`
	Seed               int64   `json:"seed"`
	BatchCount         int     `json:"batch_count"`
	AutoResizeRefImage bool    `json:"auto_resize_ref_image"`
	IncreaseRefIndex   bool    `json:"increase_ref_index"`
	ControlStrength    float64 `json:"control_strength"`
	EmbedImageMetadata bool    `json:"embed_image_metadata"`

	// Указатели для корректной передачи null в JSON
	InitImage    *string  `json:"init_image"`
	MaskImage    *string  `json:"mask_image"`
	ControlImage *string  `json:"control_image"`
	RefImages    []string `json:"ref_images"` // Срез строк для массива референсов

	SampleParams    SampleParams    `json:"sample_params"`
	Lora            []LoraParams    `json:"lora"`
	Hires           HiresConfig     `json:"hires"`
	VaeTilingParams VaeTilingParams `json:"vae_tiling_params"`

	CacheMode         string `json:"cache_mode"`
	CacheOption       string `json:"cache_option"`
	ScmMask           string `json:"scm_mask"`
	ScmPolicyDynamic  bool   `json:"scm_policy_dynamic"`
	OutputFormat      string `json:"output_format"`
	OutputCompression int    `json:"output_compression"`
}

// --- СТРУКТУРА ДЛЯ ОТВЕТА (RESPONSE) ---

type GenResponse struct {
	ID      string `json:"id"`
	Kind    string `json:"kind"`
	Status  string `json:"status"`
	Created int64  `json:"created"`
	PollURL string `json:"poll_url"`
}

// ----------------------GetStatusResponse----------------------
type JobImage struct {
	Index   int    `json:"index"`
	B64JSON string `json:"b64_json"`
}

type JobResult struct {
	OutputFormat string     `json:"output_format"`
	Images       []JobImage `json:"images,omitempty"`

	// Поля ниже пригодятся, если сервер вернет ответ для генерации видео (vid_gen)
	B64JSON  *string `json:"b64_json,omitempty"`
	MIMEType *string `json:"mime_type,omitempty"`
}

type JobError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type JobResponse struct {
	ID            string     `json:"id"`
	Kind          string     `json:"kind"`
	Status        string     `json:"status"`
	Created       int64      `json:"created"`
	Started       *int64     `json:"started"`   // *int64, так как может быть null на этапе очереди
	Completed     *int64     `json:"completed"` // *int64, так как может быть null до завершения
	QueuePosition int        `json:"queue_position"`
	Result        *JobResult `json:"result"` // *JobResult — будет nil при ошибке или в процессе работы
	Error         *JobError  `json:"error"`  // *JobError — будет nil при успешном выполнении
}

// -------------------------Img EXIF block----------------------------------------------------------
// Информация о генераторе
type GeneratorInfo struct {
	Commit  string `json:"commit"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Используемые модели
type ModelsInfo struct {
	DiffusionModel string `json:"diffusion_model"`
	Llm            string `json:"llm"`
	Vae            string `json:"vae"`
}

// Промпты
type PromptInfo struct {
	Negative string `json:"negative"`
	Positive string `json:"positive"`
}

// Параметры сэмплинга (шагов)
type SamplingInfo struct {
	CustomSigmas    []float64      `json:"custom_sigmas"` // Использован float64 для гибкости sigmas
	Eta             *float64       `json:"eta"`           // Указатель, так как в JSON может прийти null
	ExtraSampleArgs string         `json:"extra_sample_args"`
	FlowShift       *float64       `json:"flow_shift"` // Указатель, так как в JSON может прийти null
	Guidance        GuidanceConfig `json:"guidance"`
	ShiftedTimestep int            `json:"shifted_timestep"`
	Steps           int            `json:"steps"`
	Method          string         `json:"method"`
	Scheduler       string         `json:"scheduler"`
}

type SDCPPParams struct {
	AutoResizeRefImage bool          `json:"auto_resize_ref_image"`
	ClipSkip           int           `json:"clip_skip"`
	ControlStrength    float64       `json:"control_strength"`
	Generator          GeneratorInfo `json:"generator"`
	Height             int           `json:"height"`
	IncreaseRefIndex   bool          `json:"increase_ref_index"`
	Loras              []LoraParams  `json:"loras"`
	Mode               string        `json:"mode"`
	Models             ModelsInfo    `json:"models"`
	Prompt             PromptInfo    `json:"prompt"`
	Rng                string        `json:"rng"`
	Sampling           SamplingInfo  `json:"sampling"`
	Schema             string        `json:"schema"`
	Seed               int64         `json:"seed"` // Использован int64 для больших значений сидов
	Strength           float64       `json:"strength"`
	Width              int           `json:"width"`
}
