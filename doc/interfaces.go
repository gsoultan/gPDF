package doc

// Compile-time checks that DocumentBuilder implements all builder interfaces.
var (
	_ PageBuilder          = (*DocumentBuilder)(nil)
	_ MetadataBuilder      = (*DocumentBuilder)(nil)
	_ OutlineBuilder       = (*DocumentBuilder)(nil)
	_ TextBuilder          = (*DocumentBuilder)(nil)
	_ ImageBuilder         = (*DocumentBuilder)(nil)
	_ TaggedContentBuilder = (*DocumentBuilder)(nil)
	_ TaggedImageBuilder   = (*DocumentBuilder)(nil)
	_ TableBuilderAPI      = (*DocumentBuilder)(nil)
	_ FormBuilder          = (*DocumentBuilder)(nil)
	_ LayerBuilder         = (*DocumentBuilder)(nil)
)
