/** GiForm 内置 input 默认 maxlength 仅 20，未显式设置时需自行覆盖 */

export const DEFAULT_INPUT_MAXLENGTH = 255
export const DEFAULT_TEXTAREA_MAXLENGTH = 500

type FieldProps = Record<string, unknown>

/** input：未传 maxlength 时默认 255；props 中显式设置优先 */
export function inputProps(props?: FieldProps, defaultMax = DEFAULT_INPUT_MAXLENGTH): FieldProps {
  return { maxlength: defaultMax, ...props }
}

/** textarea：未传 maxlength 时默认 500；props 中显式设置优先 */
export function textareaProps(props?: FieldProps, defaultMax = DEFAULT_TEXTAREA_MAXLENGTH): FieldProps {
  return { maxlength: defaultMax, showWordLimit: true, ...props }
}
