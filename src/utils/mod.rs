use once_cell::sync::Lazy;
use regex::Regex;

// Matches XX:XX:XX:XX:XX:XX or XX-XX-XX-XX-XX-XX (separator must be consistent)
static MAC_REGEX: Lazy<Regex> = Lazy::new(|| {
    Regex::new(r"(?i)^([0-9A-F]{2}:){5}[0-9A-F]{2}$|^([0-9A-F]{2}-){5}[0-9A-F]{2}$").unwrap()
});

pub fn is_valid_mac(mac: &str) -> bool {
    MAC_REGEX.is_match(mac)
}

/// Normalizes a MAC to uppercase colon-separated form.
/// e.g. `aa-bb-cc-dd-ee-ff` → `AA:BB:CC:DD:EE:FF`
pub fn normalize_mac(mac: &str) -> String {
    mac.to_uppercase().replace('-', ":")
}

#[cfg(test)]
mod tests {
    use super::*;

    // --- is_valid_mac ---

    #[test]
    fn test_valid_mac_colon_uppercase() {
        assert!(is_valid_mac("AA:BB:CC:DD:EE:FF"));
    }

    #[test]
    fn test_valid_mac_colon_lowercase() {
        assert!(is_valid_mac("aa:bb:cc:dd:ee:ff"));
    }

    #[test]
    fn test_valid_mac_dash_uppercase() {
        assert!(is_valid_mac("AA-BB-CC-DD-EE-FF"));
    }

    #[test]
    fn test_valid_mac_dash_lowercase() {
        assert!(is_valid_mac("aa-bb-cc-dd-ee-ff"));
    }

    #[test]
    fn test_valid_mac_mixed_case() {
        assert!(is_valid_mac("aA:Bb:cC:Dd:eE:fF"));
    }

    #[test]
    fn test_invalid_mac_too_short() {
        assert!(!is_valid_mac("AA:BB:CC:DD:EE"));
    }

    #[test]
    fn test_invalid_mac_too_long() {
        assert!(!is_valid_mac("AA:BB:CC:DD:EE:FF:00"));
    }

    #[test]
    fn test_invalid_mac_mixed_separators() {
        assert!(!is_valid_mac("AA:BB-CC:DD-EE:FF"));
    }

    #[test]
    fn test_invalid_mac_no_separator() {
        assert!(!is_valid_mac("AABBCCDDEEFF"));
    }

    #[test]
    fn test_invalid_mac_invalid_chars() {
        assert!(!is_valid_mac("GG:HH:II:JJ:KK:LL"));
    }

    #[test]
    fn test_invalid_mac_empty() {
        assert!(!is_valid_mac(""));
    }

    #[test]
    fn test_invalid_mac_non_ascii_chars() {
        assert!(!is_valid_mac("éà:ùû:çñ:ää:öü:ßœ"));
    }

    // --- normalize_mac ---

    #[test]
    fn test_normalize_mac_colon_lowercase() {
        assert_eq!(normalize_mac("aa:bb:cc:dd:ee:ff"), "AA:BB:CC:DD:EE:FF");
    }

    #[test]
    fn test_normalize_mac_dash_to_colon() {
        assert_eq!(normalize_mac("aa-bb-cc-dd-ee-ff"), "AA:BB:CC:DD:EE:FF");
    }

    #[test]
    fn test_normalize_mac_already_normalized() {
        assert_eq!(normalize_mac("AA:BB:CC:DD:EE:FF"), "AA:BB:CC:DD:EE:FF");
    }

    #[test]
    fn test_normalize_mac_mixed_case_dash() {
        assert_eq!(normalize_mac("aA-bB-cC-dD-eE-fF"), "AA:BB:CC:DD:EE:FF");
    }
}
