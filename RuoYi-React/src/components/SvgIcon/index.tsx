// SvgIcon —— 对位基准 SvgIcon 组件（svg 精灵 use 引用）

interface SvgIconProps {
  iconClass: string
  className?: string
  color?: string
  size?: string | number
}

export default function SvgIcon({ iconClass, className, color, size = '1em' }: SvgIconProps) {
  return (
    <svg
      className={className}
      // 对位基准:width/height 1em（随字号缩放）、vertical-align:-2px、fill currentColor
      style={{ color, width: size, height: size, verticalAlign: '-0.15em', fill: 'currentColor' }}
      aria-hidden="true"
    >
      <use href={`#icon-${iconClass}`} />
    </svg>
  )
}
