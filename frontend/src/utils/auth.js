// Utility function to clear all cookies and redirect to home
export const clearCookiesAndRedirectHome = () => {
  // Clear all cookies by setting them to expire in the past
  document.cookie.split(";").forEach((cookie) => {
    const eqPos = cookie.indexOf("=");
    const name = eqPos > -1 ? cookie.substr(0, eqPos).trim() : cookie.trim();
    // Clear cookie for current domain
    document.cookie = `${name}=;expires=Thu, 01 Jan 1970 00:00:00 GMT;path=/`;
    // Clear cookie for parent domain (if applicable)
    document.cookie = `${name}=;expires=Thu, 01 Jan 1970 00:00:00 GMT;path=/;domain=${window.location.hostname}`;
    // Clear cookie for subdomain wildcard (if applicable)
    document.cookie = `${name}=;expires=Thu, 01 Jan 1970 00:00:00 GMT;path=/;domain=.${window.location.hostname}`;
  });
  
  // Clear localStorage and sessionStorage
  localStorage.clear();
  sessionStorage.clear();
  
  // Redirect to home page
  window.location.href = '/';
};

// Check if user has any authentication-related cookies
const hasAuthCookies = () => {
  return document.cookie.includes('auth-session');
};

// Redirect to login page
export const redirectToLogin = () => {
  window.location.href = '/login';
};

// Check if user is authenticated by calling the API
export const checkAuthentication = async () => {
  try {
    const response = await fetch('/api/user', { 
      credentials: 'include',
      method: 'GET'
    });
    
    if (response.status === 401 || response.status === 403) {
      // Authentication failed - check if user had cookies
      if (hasAuthCookies()) {
        // User had cookies but they're invalid/expired - clear and go to home
        clearCookiesAndRedirectHome();
      } else {
        // User has no cookies - send to login
        redirectToLogin();
      }
      return false;
    }
    
    if (!response.ok) {
      throw new Error('Network response was not ok');
    }
    
    return await response.json();
  } catch (error) {
    console.error('Authentication check failed:', error);
    // If there's a network error, check for cookies to decide where to redirect
    if (hasAuthCookies()) {
      // User had cookies but there's an error - clear and go to home
      clearCookiesAndRedirectHome();
    } else {
      // User has no cookies - send to login
      redirectToLogin();
    }
    return false;
  }
};
