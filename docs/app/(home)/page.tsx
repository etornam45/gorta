import Image from 'next/image';
import Link from 'next/link';
import {
  ArrowRight,
  Check,
  Cookie,
  Database,
  KeyRound,
  Link2,
  Mail,
  Plug,
  Shield,
} from 'lucide-react';
import { appName, docsRoute, gitConfig } from '@/lib/shared';
import { ServerCodeBlock } from 'fumadocs-ui/components/codeblock.rsc';

const githubUrl = `https://github.com/${gitConfig.user}/${gitConfig.repo}`;

const pillars = [
  {
    icon: Database,
    title: 'You own the data',
    description:
      'Implement core.Storage against your database, ORM, and schema. Gorta never opens a connection or runs migrations for you.',
  },
  {
    icon: Plug,
    title: 'Composable plugins',
    description:
      'Enable email/password, magic links, OAuth, and more. Each plugin registers HTTP routes and optional storage interfaces.',
  },
  {
    icon: Cookie,
    title: 'Sessions & cookies',
    description:
      'Password hashing, secure cookies, session validation, and RequireAuth middleware — without locking you into a framework.',
  },
] as const;

const availableAuth = [
  { icon: Mail, label: 'Email & password' },
  { icon: Link2, label: 'Magic links' },
  { icon: KeyRound, label: 'Google & GitHub OAuth' },
] as const;

const plannedAuth = [
  'Passkeys (WebAuthn)',
  'Email OTP',
  'Phone number',
  'Username',
  'Anonymous sign-in',
  'One-tap sign-in',
] as const;

const quickStartCode = `storage := sqladapter.New(db)

emailPlugin, _ := emailpassword.New(storage, storage, mailer, emailpassword.Config{
    VerifyEmail: true,
})

auth, _ := gorta.New(storage, gorta.Config{
    Secret:  os.Getenv("SECRET"),
    BaseURL: os.Getenv("BASE_URL"),
}, gorta.WithPlugin(emailPlugin))

mux := http.NewServeMux()
mux.Handle("/auth/", http.StripPrefix("/auth", auth.Handler()))
mux.Handle("/dashboard", auth.RequireAuth()(http.HandlerFunc(dashboard)))

http.ListenAndServe(":8080", auth.Middleware()(mux))`;

export default async function HomePage() {
  return (
    <div className="flex flex-1 flex-col">
      {/* Hero */}
      <section className="border-b border-fd-border bg-fd-background px-6 py-20 md:py-28">
        <div className="mx-auto flex max-w-4xl flex-col items-center text-center">
          <Image
            src="/gorta.png"
            alt=""
            width={88}
            height={88}
            className="mb-6 rounded-2xl"
            priority
          />
          <p className="mb-3 text-sm font-medium tracking-wide text-fd-muted-foreground uppercase">
            Authentication for Go
          </p>
          <h1 className="text-4xl font-bold tracking-tight text-fd-foreground md:text-5xl">
            {appName}
          </h1>
          <p className="mt-5 max-w-2xl text-lg leading-relaxed text-fd-muted-foreground">
            A minimal, adapter-based auth library. Gorta handles password hashing,
            sessions, OAuth flows, and secure cookies — you implement storage against
            your own database.
          </p>
          <div className="mt-10 flex flex-wrap items-center justify-center gap-3">
            <Link
              href={docsRoute}
              className="inline-flex items-center gap-2 rounded-lg bg-fd-primary px-5 py-2.5 text-sm font-medium text-fd-primary-foreground transition-opacity hover:opacity-90"
            >
              Get started
              <ArrowRight className="size-4" aria-hidden />
            </Link>
            <a
              href={githubUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-2 rounded-lg border border-fd-border bg-fd-card px-5 py-2.5 text-sm font-medium text-fd-foreground transition-colors hover:bg-fd-accent"
            >
              View on GitHub
            </a>
          </div>
        </div>
      </section>


      {/* Pillars */}
      <section className="border-y border-fd-border bg-fd-card/50 px-6 py-16">
        <div className="mx-auto max-w-5xl">
          <h2 className="text-center text-2xl font-semibold text-fd-foreground">
            Built for control, not ceremony
          </h2>
          <p className="mx-auto mt-3 max-w-xl text-center text-fd-muted-foreground">
            Standard library HTTP, explicit configuration, no ORM requirement.
          </p>
          <ul className="mt-12 grid gap-8 md:grid-cols-3">
            {pillars.map(({ icon: Icon, title, description }) => (
              <li
                key={title}
                className="rounded-xl border border-fd-border bg-fd-background p-6"
              >
                <div className="mb-4 flex size-10 items-center justify-center rounded-lg bg-fd-primary/10 text-fd-primary">
                  <Icon className="size-5" aria-hidden />
                </div>
                <h3 className="font-semibold text-fd-foreground">{title}</h3>
                <p className="mt-2 text-sm leading-relaxed text-fd-muted-foreground">
                  {description}
                </p>
              </li>
            ))}
          </ul>
        </div>
      </section>

      {/* Code */}
      <section className="px-6 py-16">
        <div className="mx-auto max-w-3xl">
          <div className="flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
            <div>
              <h2 className="text-2xl font-semibold text-fd-foreground">
                A few lines to go live
              </h2>
              <p className="mt-2 text-fd-muted-foreground">
                Mount routes under{' '}
                <code className="rounded bg-fd-muted px-1.5 py-0.5 text-sm">/auth</code>,
                wrap your mux with session middleware, protect handlers with{' '}
                <code className="rounded bg-fd-muted px-1.5 py-0.5 text-sm">
                  RequireAuth
                </code>
                .
              </p>
            </div>
            <a
              href={`${githubUrl}/blob/${gitConfig.branch}/cmd/main.go`}
              target="_blank"
              rel="noopener noreferrer"
              className="shrink-0 text-sm font-medium text-fd-primary hover:underline"
            >
              Full example →
            </a>
          </div>
          <ServerCodeBlock
            code={quickStartCode}
            lang="go"
            codeblock={{ className: 'mt-6' }}
          />
        </div>
      </section>

      {/* Auth methods */}
      <section className="border-t border-fd-border bg-fd-card/50 px-6 py-16">
        <div className="mx-auto max-w-4xl">
          <h2 className="text-center text-2xl font-semibold text-fd-foreground">
            Auth methods
          </h2>
          <div className="mt-10 grid gap-10 md:grid-cols-2">
            <div>
              <h3 className="mb-4 flex items-center gap-2 text-sm font-medium text-fd-foreground uppercase tracking-wide">
                <Check className="size-4 text-fd-primary" aria-hidden />
                Available today
              </h3>
              <ul className="space-y-3">
                {availableAuth.map(({ icon: Icon, label }) => (
                  <li
                    key={label}
                    className="flex items-center gap-3 rounded-lg border border-fd-border bg-fd-background px-4 py-3 text-sm"
                  >
                    <Icon className="size-4 shrink-0 text-fd-primary" aria-hidden />
                    {label}
                  </li>
                ))}
              </ul>
            </div>
            <div>
              <h3 className="mb-4 text-sm font-medium text-fd-muted-foreground uppercase tracking-wide">
                On the roadmap
              </h3>
              <ul className="space-y-2">
                {plannedAuth.map((label) => (
                  <li
                    key={label}
                    className="flex items-center gap-2 text-sm text-fd-muted-foreground"
                  >
                    <span className="size-1.5 shrink-0 rounded-full bg-fd-border" />
                    {label}
                  </li>
                ))}
              </ul>
            </div>
          </div>
        </div>
      </section>

      {/* Bottom CTA */}
      <section className="border-t border-fd-border px-6 py-50">
        <div className="mx-auto max-w-2xl text-center">
          <h2 className="text-2xl font-semibold text-fd-foreground">
            Ready to wire up auth?
          </h2>
          <p className="mt-3 text-fd-muted-foreground">
            Follow the getting started guide, then copy patterns from{' '}
            <code className="rounded bg-fd-muted px-1.5 py-0.5 text-sm">cmd/main.go</code>.
          </p>
          <Link
            href={docsRoute}
            className="mt-8 inline-flex items-center gap-2 rounded-lg bg-fd-primary px-6 py-3 text-sm font-medium text-fd-primary-foreground transition-opacity hover:opacity-90"
          >
            Read the docs
            <ArrowRight className="size-4" aria-hidden />
          </Link>
        </div>
      </section>
    </div>
  );
}
