import { useSettings } from '../contexts/SettingsContext'

export function SettingsPanel() {
  const {
    settingsOpen,
    closeSettings,
    shepherdUrl,
    shepherdToken,
    meadowUrl,
    meadowToken,
    shepherdAuthRequired,
    meadowAuthRequired,
    setShepherdUrl,
    setShepherdTokenValue,
    setMeadowUrl,
    setMeadowTokenValue,
    refreshAuthStatus,
  } = useSettings()

  if (!settingsOpen) return null

  return (
    <div className="drawer-scrim" onClick={closeSettings}>
      <aside
        className="drawer"
        role="dialog"
        aria-labelledby="settings-title"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="drawer__head">
          <h2 id="settings-title">Settings</h2>
          <button
            type="button"
            className="btn btn--icon"
            onClick={closeSettings}
            aria-label="Close settings"
          >
            ×
          </button>
        </header>

        <div className="drawer__body">
          <section className="settings-section">
            <h3>Shepherd API</h3>
            {shepherdAuthRequired && (
              <p className="settings-hint">
                This cluster requires a Bearer token for API access.
              </p>
            )}
            <label className="field">
              <span className="field__label">Base URL</span>
              <input
                className="field__input mono"
                value={shepherdUrl}
                onChange={(e) => setShepherdUrl(e.target.value)}
              />
            </label>
            <label className="field">
              <span className="field__label">API token</span>
              <input
                className="field__input mono"
                type="password"
                value={shepherdToken}
                onChange={(e) => setShepherdTokenValue(e.target.value)}
                placeholder="Bearer token (if required)"
                autoComplete="off"
              />
            </label>
          </section>

          <section className="settings-section">
            <h3>Meadow registry</h3>
            {meadowAuthRequired && (
              <p className="settings-hint">
                Meadow requires a Bearer token for API access.
              </p>
            )}
            <label className="field">
              <span className="field__label">Base URL</span>
              <input
                className="field__input mono"
                value={meadowUrl}
                onChange={(e) => setMeadowUrl(e.target.value)}
              />
            </label>
            <label className="field">
              <span className="field__label">API token</span>
              <input
                className="field__input mono"
                type="password"
                value={meadowToken}
                onChange={(e) => setMeadowTokenValue(e.target.value)}
                placeholder="Bearer token (if required)"
                autoComplete="off"
              />
            </label>
          </section>
        </div>

        <footer className="drawer__foot">
          <button type="button" className="btn" onClick={refreshAuthStatus}>
            Test connection
          </button>
          <button type="button" className="btn btn--primary" onClick={closeSettings}>
            Done
          </button>
        </footer>
      </aside>
    </div>
  )
}
