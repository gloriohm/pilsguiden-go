// tailwind.config.js
export default {
  darkMode: ['class'],
  content: ['./**/*.{html,templ,go,ts,tsx,js}'],
  theme: {
    extend: {
      colors: {
        bg: 'rgb(var(--bg) / <alpha-value>)',
        fg: 'rgb(var(--fg) / <alpha-value>)',
        muted: 'rgb(var(--muted) / <alpha-value>)',
        'muted-fg': 'rgb(var(--muted-fg) / <alpha-value>)',
        saturated: 'rgb(var(--sat) / <alpha-value',
        border: 'rgb(var(--border) / <alpha-value>)',
        ring: 'rgb(var(--ring) / <alpha-value>)',
        primary: 'rgb(var(--primary) / <alpha-value>)',
        'primary-foreground': 'rgb(var(--primary-foreground) / <alpha-value>)',
        success: 'rgb(var(--success) / <alpha-value>)',
        warning: 'rgb(var(--warning) / <alpha-value>)',
        danger: 'rgb(var(--danger) / <alpha-value>)',
      },
      borderColor: {
        DEFAULT: 'rgb(var(--border) / <alpha-value>)',
      },
      textColor: {
        skin: {
          base: 'rgb(var(--fg) / <alpha-value>)',
          muted: 'rgb(var(--muted-fg) / <alpha-value>)',
          onPrimary: 'rgb(var(--primary-foreground) / <alpha-value>)',
        },
      },
      backgroundColor: {
        skin: {
          base: 'rgb(var(--bg) / <alpha-value>)',
          muted: 'rgb(var(--muted) / <alpha-value>)',
          saturated: 'rgb(var(--sat) / <alpha-value',
          primary: 'rgb(var(--primary) / <alpha-value>)',
        },
      },
    },
  },
  plugins: [],
}
