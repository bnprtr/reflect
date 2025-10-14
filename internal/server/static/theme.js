// Theme management for light/dark mode
(function() {
  'use strict';
  
  const THEME_KEY = 'reflect-theme';
  const THEME_LIGHT = 'light';
  const THEME_DARK = 'dark';
  
  // Get current theme from localStorage or system preference
  function getCurrentTheme() {
    const stored = localStorage.getItem(THEME_KEY);
    if (stored === THEME_LIGHT || stored === THEME_DARK) {
      return stored;
    }
    
    // Fall back to system preference
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? THEME_DARK : THEME_LIGHT;
  }
  
  // Apply theme to document
  function applyTheme(theme) {
    if (theme === THEME_DARK) {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  }
  
  // Update theme toggle button
  function updateToggleButton(theme) {
    const toggle = document.getElementById('theme-toggle');
    if (!toggle) return;
    
    const icon = toggle.querySelector('svg');
    if (!icon) return;
    
    if (theme === THEME_DARK) {
      // Show sun icon for switching to light mode
      icon.innerHTML = `
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
      `;
      toggle.setAttribute('aria-label', 'Switch to light mode');
    } else {
      // Show moon icon for switching to dark mode
      icon.innerHTML = `
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
      `;
      toggle.setAttribute('aria-label', 'Switch to dark mode');
    }
  }
  
  // Toggle theme
  function toggleTheme() {
    const current = getCurrentTheme();
    const newTheme = current === THEME_LIGHT ? THEME_DARK : THEME_LIGHT;
    
    localStorage.setItem(THEME_KEY, newTheme);
    applyTheme(newTheme);
    updateToggleButton(newTheme);
  }
  
  // Initialize theme on page load
  function initTheme() {
    const theme = getCurrentTheme();
    applyTheme(theme);
    updateToggleButton(theme);
    
    // Add click handler to toggle button
    const toggle = document.getElementById('theme-toggle');
    if (toggle) {
      toggle.addEventListener('click', toggleTheme);
    }
    
    // Listen for system theme changes
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', function(e) {
      // Only update if no explicit preference is stored
      if (!localStorage.getItem(THEME_KEY)) {
        const newTheme = e.matches ? THEME_DARK : THEME_LIGHT;
        applyTheme(newTheme);
        updateToggleButton(newTheme);
      }
    });
  }
  
  // Initialize when DOM is ready
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initTheme);
  } else {
    initTheme();
  }
})();
