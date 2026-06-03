CREATE TABLE tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    type VARCHAR(20) NOT NULL CHECK (type IN ('usage', 'style', 'category')),
    usage_count INT NOT NULL DEFAULT 0
);

CREATE TABLE icon_tags (
    icon_id UUID NOT NULL REFERENCES icons(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (icon_id, tag_id)
);

CREATE TABLE icon_colors (
    icon_id UUID NOT NULL REFERENCES icons(id) ON DELETE CASCADE,
    color_hex VARCHAR(7) NOT NULL,
    role VARCHAR(10) NOT NULL DEFAULT 'fill',
    PRIMARY KEY (icon_id, color_hex)
);

CREATE TABLE icon_themes (
    icon_id UUID NOT NULL REFERENCES icons(id) ON DELETE CASCADE,
    theme_name VARCHAR(100) NOT NULL,
    PRIMARY KEY (icon_id, theme_name)
);

-- Helper: extract HSL hue from hex color for color range filtering
CREATE OR REPLACE FUNCTION hue_from_hex(hex_color VARCHAR) RETURNS INT AS $$
DECLARE
    r FLOAT; g FLOAT; b FLOAT;
    mx FLOAT; mn FLOAT; delta FLOAT;
    h FLOAT;
BEGIN
    r := ('x' || substring(hex_color FROM 2 FOR 2))::int / 255.0;
    g := ('x' || substring(hex_color FROM 4 FOR 2))::int / 255.0;
    b := ('x' || substring(hex_color FROM 6 FOR 2))::int / 255.0;
    mx := GREATEST(r, g, b);
    mn := LEAST(r, g, b);
    delta := mx - mn;
    IF delta = 0 THEN
        RETURN 0;
    END IF;
    IF mx = r THEN
        h := 60 * (mod((g - b) / delta, 6));
    ELSIF mx = g THEN
        h := 60 * ((b - r) / delta + 2);
    ELSE
        h := 60 * ((r - g) / delta + 4);
    END IF;
    IF h < 0 THEN
        h := h + 360;
    END IF;
    RETURN round(h)::int;
END;
$$ LANGUAGE plpgsql IMMUTABLE;
