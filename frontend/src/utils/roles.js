export const ROLES = {
  ADMIN: "admin",
  USER: "user",
};

/**
 * Check if the user's role is allowed for the given route roles.
 * @param {string[] | undefined} requiredRoles - Array of allowed roles, or undefined (no restriction).
 * @param {string} userRole - The user's role.
 * @returns {boolean}
 */
export function canAccessRole(requiredRoles, userRole) {
  if (!requiredRoles || requiredRoles.length === 0) {
    return true;
  }

  return requiredRoles.includes(userRole);
}

/**
 * Get the default admin path for a given role.
 * @param {string} role - The user's role.
 * @returns {string}
 */
export function getDefaultPath(role) {
  return "/";
}