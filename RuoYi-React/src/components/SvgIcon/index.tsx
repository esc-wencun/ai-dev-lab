// SvgIcon —— 对位基准 SvgIcon 组件（svg 精灵 use 引用）

interface SvgIconProps {
  iconClass: string
  className?: string
  color?: string
  size?: string | number
}

export default function SvgIcon({ iconClass, className, color, size = 14 }: SvgIconProps) {
  return (
    <svg
      className={className}
      style={{ color, width: size, height: size, verticalAlign: 'middle', fill: 'currentColor' }}
      aria-hidden="true"
    >
      <use href={`#icon-${iconClass}`} />
    </svg>
  )
}
