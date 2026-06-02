import { useEffect, useState, type ReactNode } from "react";
import {
  loadProfile,
  openInNeovim,
  saveProfile,
  validateProfile,
  validateSSHPublicKey,
  type ValidationResult,
} from "../api/backend";
import {
  createDefaultProfile,
  createDefaultUser,
  type Profile,
  type Script,
  type User,
} from "../types/profile";

const pendingValidation: ValidationResult = {
  valid: false,
  problems: ["Waiting for backend validation."],
};

function ProfileEditor() {
  const [profile, setProfile] = useState(createDefaultProfile);
  const [path, setPath] = useState("");
  const [validation, setValidation] = useState(pendingValidation);
  const [message, setMessage] = useState("Create a profile or load an existing YAML file.");

  useEffect(() => {
    let active = true;
    const timer = window.setTimeout(() => {
      void validateProfile(profile).then((result) => {
        if (active) {
          setValidation(result);
        }
      });
    }, 150);

    return () => {
      active = false;
      window.clearTimeout(timer);
    };
  }, [profile]);

  async function handleLoad() {
    try {
      const loaded = await loadProfile(path);
      setProfile(normalizeProfile(loaded));
      setMessage(`Loaded ${path}`);
    } catch (error) {
      setMessage(errorMessage(error));
    }
  }

  async function handleSave() {
    try {
      await saveProfile(path, profile);
      setMessage(`Saved ${path}`);
    } catch (error) {
      setMessage(errorMessage(error));
    }
  }

  async function handleNeovim() {
    try {
      await openInNeovim(path);
      setMessage(`Opened ${path} in Neovim`);
    } catch (error) {
      setMessage(errorMessage(error));
    }
  }

  return (
    <div className="profile-editor">
      <section className="editor-toolbar panel">
        <div>
          <span className="panel-label">Profile YAML</span>
          <p>{message}</p>
        </div>
        <label className="path-field">
          <span>File path</span>
          <input
            onChange={(event) => setPath(event.target.value)}
            placeholder="/path/to/profile.yaml"
            value={path}
          />
        </label>
        <div className="toolbar-actions">
          <button className="secondary-action" onClick={() => setProfile(createDefaultProfile())} type="button">
            New
          </button>
          <button className="secondary-action" disabled={!path} onClick={handleLoad} type="button">
            Load
          </button>
          <button className="secondary-action" disabled={!path} onClick={handleNeovim} type="button">
            Open in Neovim
          </button>
          <button className="primary-action" disabled={!path || !validation.valid} onClick={handleSave} type="button">
            Save
          </button>
        </div>
      </section>

      <ValidationSummary result={validation} />

      <div className="form-grid">
        <EditorSection title="Profile">
          <Field label="Name">
            <input value={profile.name} onChange={(event) => setProfile({ ...profile, name: event.target.value })} />
          </Field>
        </EditorSection>

        <EditorSection title="Operating System">
          <Field label="Family">
            <select
              value={profile.os.family}
              onChange={(event) => setProfile({ ...profile, os: { ...profile.os, family: event.target.value } })}
            >
              {["alpine", "ubuntu", "debian", "rocky", "fedora", "arch", "opensuse"].map((family) => (
                <option key={family}>{family}</option>
              ))}
            </select>
          </Field>
          <Field label="Version">
            <input
              value={profile.os.version}
              onChange={(event) => setProfile({ ...profile, os: { ...profile.os, version: event.target.value } })}
            />
          </Field>
          <Field label="Architecture">
            <select
              value={profile.os.architecture}
              onChange={(event) => setProfile({ ...profile, os: { ...profile.os, architecture: event.target.value } })}
            >
              {["x86_64", "aarch64", "arm64", "armv7", "riscv64"].map((architecture) => (
                <option key={architecture}>{architecture}</option>
              ))}
            </select>
          </Field>
        </EditorSection>

        <EditorSection title="System">
          <Field label="Hostname">
            <input
              value={profile.system.hostname}
              onChange={(event) =>
                setProfile({ ...profile, system: { ...profile.system, hostname: event.target.value } })
              }
            />
          </Field>
          <Field label="Timezone">
            <input
              value={profile.system.timezone}
              onChange={(event) =>
                setProfile({ ...profile, system: { ...profile.system, timezone: event.target.value } })
              }
            />
          </Field>
          <Field label="Keyboard">
            <input
              value={profile.system.keyboard}
              onChange={(event) =>
                setProfile({ ...profile, system: { ...profile.system, keyboard: event.target.value } })
              }
            />
          </Field>
          <Field label="Locale">
            <input
              value={profile.system.locale}
              onChange={(event) => setProfile({ ...profile, system: { ...profile.system, locale: event.target.value } })}
            />
          </Field>
        </EditorSection>

        <EditorSection title="Network">
          <Field label="Mode">
            <select
              value={profile.network.mode}
              onChange={(event) =>
                setProfile({ ...profile, network: { ...profile.network, mode: event.target.value } })
              }
            >
              <option>dhcp</option>
              <option>static</option>
            </select>
          </Field>
          <Field label="Interface">
            <input
              value={profile.network.interface}
              onChange={(event) =>
                setProfile({ ...profile, network: { ...profile.network, interface: event.target.value } })
              }
            />
          </Field>
          {profile.network.mode === "static" && (
            <>
              <Field label="Address (CIDR)">
                <input
                  value={profile.network.address}
                  onChange={(event) =>
                    setProfile({ ...profile, network: { ...profile.network, address: event.target.value } })
                  }
                />
              </Field>
              <Field label="Gateway">
                <input
                  value={profile.network.gateway}
                  onChange={(event) =>
                    setProfile({ ...profile, network: { ...profile.network, gateway: event.target.value } })
                  }
                />
              </Field>
            </>
          )}
          <Field label="DNS servers">
            <input
              value={joinList(profile.network.dns)}
              onChange={(event) =>
                setProfile({ ...profile, network: { ...profile.network, dns: splitList(event.target.value) } })
              }
            />
          </Field>
        </EditorSection>

        <EditorSection title="Disk">
          <Field label="Mode">
            <select
              value={profile.disk.mode}
              onChange={(event) => setProfile({ ...profile, disk: { ...profile.disk, mode: event.target.value } })}
            >
              <option>wipe</option>
              <option>preserve</option>
              <option>manual</option>
            </select>
          </Field>
          <Field label="Target">
            <input
              value={profile.disk.target}
              onChange={(event) => setProfile({ ...profile, disk: { ...profile.disk, target: event.target.value } })}
            />
          </Field>
          <Field label="Install mode">
            <input
              value={profile.disk.install_mode}
              onChange={(event) =>
                setProfile({ ...profile, disk: { ...profile.disk, install_mode: event.target.value } })
              }
            />
          </Field>
          <Field label="Filesystem">
            <select
              value={profile.disk.filesystem}
              onChange={(event) =>
                setProfile({ ...profile, disk: { ...profile.disk, filesystem: event.target.value } })
              }
            >
              {["ext4", "xfs", "btrfs", "f2fs"].map((filesystem) => (
                <option key={filesystem}>{filesystem}</option>
              ))}
            </select>
          </Field>
          <Checkbox
            checked={profile.disk.confirm_destructive}
            label="Confirm destructive installer disk mode"
            onChange={(checked) => setProfile({ ...profile, disk: { ...profile.disk, confirm_destructive: checked } })}
          />
        </EditorSection>

        <EditorSection title="SSH">
          <Checkbox
            checked={profile.ssh.enabled}
            label="Enable SSH"
            onChange={(checked) => setProfile({ ...profile, ssh: { ...profile.ssh, enabled: checked } })}
          />
          <Checkbox
            checked={profile.ssh.permit_root_login}
            label="Permit root login"
            onChange={(checked) => setProfile({ ...profile, ssh: { ...profile.ssh, permit_root_login: checked } })}
          />
          <Checkbox
            checked={profile.ssh.password_authentication}
            label="Permit password authentication"
            onChange={(checked) =>
              setProfile({ ...profile, ssh: { ...profile.ssh, password_authentication: checked } })
            }
          />
        </EditorSection>

        <EditorSection title="Packages and Services">
          <Field label="Package presets">
            <input
              value={joinList(profile.packages.presets)}
              onChange={(event) =>
                setProfile({ ...profile, packages: { ...profile.packages, presets: splitList(event.target.value) } })
              }
            />
          </Field>
          <Field label="Explicit packages">
            <input
              value={joinList(profile.packages.names)}
              onChange={(event) =>
                setProfile({ ...profile, packages: { ...profile.packages, names: splitList(event.target.value) } })
              }
            />
          </Field>
          <Field label="Enabled services">
            <input
              value={joinList(profile.services.enable)}
              onChange={(event) =>
                setProfile({ ...profile, services: { enable: splitList(event.target.value) } })
              }
            />
          </Field>
        </EditorSection>
      </div>

      <UsersEditor profile={profile} setProfile={setProfile} />
      <ScriptsEditor path={path} profile={profile} setMessage={setMessage} setProfile={setProfile} />
    </div>
  );
}

function UsersEditor({ profile, setProfile }: ProfileEditorProps) {
  function updateUser(index: number, patch: Partial<User>) {
    const users = profile.users.map((user, userIndex) => (userIndex === index ? { ...user, ...patch } : user));
    setProfile({ ...profile, users });
  }

  return (
    <section className="panel editor-list">
      <div className="section-header">
        <span className="panel-label">Users and SSH Keys</span>
        <button
          className="secondary-action"
          onClick={() => setProfile({ ...profile, users: [...profile.users, createDefaultUser()] })}
          type="button"
        >
          Add User
        </button>
      </div>
      {profile.users.map((user, index) => (
        <div className="list-card" key={`${index}-${user.name}`}>
          <div className="list-card-grid">
            <Field label="User name">
              <input value={user.name} onChange={(event) => updateUser(index, { name: event.target.value })} />
            </Field>
            <Field label="Shell">
              <input value={user.shell} onChange={(event) => updateUser(index, { shell: event.target.value })} />
            </Field>
            <Field label="Groups">
              <input
                value={joinList(user.groups)}
                onChange={(event) => updateUser(index, { groups: splitList(event.target.value) })}
              />
            </Field>
            <Checkbox checked={user.sudo} label="Enable sudo" onChange={(checked) => updateUser(index, { sudo: checked })} />
          </div>
          <SSHKeysEditor keys={user.ssh_keys} onChange={(ssh_keys) => updateUser(index, { ssh_keys })} />
          <button
            className="text-action danger"
            disabled={profile.users.length === 1}
            onClick={() => setProfile({ ...profile, users: profile.users.filter((_, userIndex) => userIndex !== index) })}
            type="button"
          >
            Remove user
          </button>
        </div>
      ))}
    </section>
  );
}

function SSHKeysEditor({ keys, onChange }: { keys: string[]; onChange: (keys: string[]) => void }) {
  const [problems, setProblems] = useState<string[]>([]);

  useEffect(() => {
    let active = true;
    const timer = window.setTimeout(() => {
      void Promise.all(keys.map((key) => validateSSHPublicKey(key))).then((results) => {
        if (active) {
          setProblems(results.flatMap((result) => result.problems));
        }
      });
    }, 150);
    return () => {
      active = false;
      window.clearTimeout(timer);
    };
  }, [keys]);

  return (
    <Field label="SSH public keys, one per line">
      <textarea value={keys.join("\n")} onChange={(event) => onChange(splitLines(event.target.value))} />
      {problems.map((problem) => (
        <span className="inline-error" key={problem}>
          {problem}
        </span>
      ))}
    </Field>
  );
}

function ScriptsEditor({
  path,
  profile,
  setMessage,
  setProfile,
}: ProfileEditorProps & { path: string; setMessage: (message: string) => void }) {
  function updateScript(index: number, patch: Partial<Script>) {
    const scripts = profile.post_install.scripts.map((script, scriptIndex) =>
      scriptIndex === index ? { ...script, ...patch } : script,
    );
    setProfile({ ...profile, post_install: { scripts } });
  }

  async function handleNeovim() {
    try {
      await openInNeovim(path);
      setMessage(`Opened ${path} in Neovim for post-install script editing`);
    } catch (error) {
      setMessage(errorMessage(error));
    }
  }

  return (
    <section className="panel editor-list">
      <div className="section-header">
        <span className="panel-label">Post-Install Scripts</span>
        <div className="toolbar-actions">
          <button className="secondary-action" disabled={!path} onClick={handleNeovim} type="button">
            Edit scripts in Neovim
          </button>
          <button
            className="secondary-action"
            onClick={() =>
              setProfile({
                ...profile,
                post_install: {
                  scripts: [...profile.post_install.scripts, { name: "script.sh", content: "#!/bin/sh\n" }],
                },
              })
            }
            type="button"
          >
            Add Script
          </button>
        </div>
      </div>
      {profile.post_install.scripts.length === 0 && <p>No post-install scripts configured.</p>}
      {profile.post_install.scripts.map((script, index) => (
        <div className="list-card" key={`${index}-${script.name}`}>
          <Field label="File name">
            <input value={script.name} onChange={(event) => updateScript(index, { name: event.target.value })} />
          </Field>
          <Field label="Script content">
            <textarea
              className="script-editor"
              value={script.content}
              onChange={(event) => updateScript(index, { content: event.target.value })}
            />
          </Field>
          <button
            className="text-action danger"
            onClick={() =>
              setProfile({
                ...profile,
                post_install: {
                  scripts: profile.post_install.scripts.filter((_, scriptIndex) => scriptIndex !== index),
                },
              })
            }
            type="button"
          >
            Remove script
          </button>
        </div>
      ))}
    </section>
  );
}

function ValidationSummary({ result }: { result: ValidationResult }) {
  return (
    <section className={result.valid ? "validation-summary valid" : "validation-summary invalid"}>
      <strong>{result.valid ? "Profile is valid" : `${result.problems.length} validation issue(s)`}</strong>
      {!result.valid && (
        <ul>
          {result.problems.map((problem) => (
            <li key={problem}>{problem}</li>
          ))}
        </ul>
      )}
    </section>
  );
}

function EditorSection({ children, title }: { children: ReactNode; title: string }) {
  return (
    <section className="panel form-section">
      <span className="panel-label">{title}</span>
      <div className="field-grid">{children}</div>
    </section>
  );
}

function Field({ children, label }: { children: ReactNode; label: string }) {
  return (
    <label className="field">
      <span>{label}</span>
      {children}
    </label>
  );
}

function Checkbox({ checked, label, onChange }: { checked: boolean; label: string; onChange: (checked: boolean) => void }) {
  return (
    <label className="checkbox-field">
      <input checked={checked} onChange={(event) => onChange(event.target.checked)} type="checkbox" />
      <span>{label}</span>
    </label>
  );
}

function normalizeProfile(profile: Profile): Profile {
  return {
    ...profile,
    network: { ...profile.network, dns: profile.network.dns ?? [] },
    packages: { presets: profile.packages?.presets ?? [], names: profile.packages?.names ?? [] },
    services: { enable: profile.services?.enable ?? [] },
    validation: {
      ssh: profile.validation?.ssh ?? { enabled: false, user: "" },
      commands: profile.validation?.commands ?? [],
    },
    post_install: { scripts: profile.post_install?.scripts ?? [] },
    os_specific: profile.os_specific ?? {},
    users: profile.users.map((user) => ({ ...user, groups: user.groups ?? [], ssh_keys: user.ssh_keys ?? [] })),
  };
}

function splitList(value: string): string[] {
  return value
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean);
}

function splitLines(value: string): string[] {
  return value
    .split("\n")
    .map((item) => item.trim())
    .filter(Boolean);
}

function joinList(values: string[]): string {
  return values.join(", ");
}

function errorMessage(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}

interface ProfileEditorProps {
  profile: Profile;
  setProfile: (profile: Profile) => void;
}

export default ProfileEditor;
