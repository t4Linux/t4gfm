package utils

const CntFooterPanels = 3

const BorderPaddingForFooter = 2

// Including borders
func FullFooterHeight(footerHeight int, toggleFooter bool) int {
	if toggleFooter {
		if footerHeight <= 1 {
			return footerHeight
		}
		return footerHeight + BorderPaddingForFooter
	}
	return 0
}
