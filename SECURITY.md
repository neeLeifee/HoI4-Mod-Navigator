# Security policy

This project is a preview. The native COM DLL runs inside Windows Explorer; defects can cause Explorer instability. No external security audit or code signing is claimed. Use the compatibility mode if necessary.

Please report suspected vulnerabilities privately to the maintainer via the repository's Security → Report a vulnerability feature, if enabled, rather than posting exploit details in public issues. Do not attach private settings, account data, or full personal paths. If private reporting is unavailable, ask for a private contact without publishing technical details.

Reports should include the version, Windows version, affected feature, minimal reproduction using test directories, and expected versus actual behavior. Do not test against other people's computers or data.

The application validates configured roots and rejects reparse points and out-of-root paths; this is not a security boundary against an adversary with write access to the same Windows user account. Review changes to registry and filesystem operations carefully.
