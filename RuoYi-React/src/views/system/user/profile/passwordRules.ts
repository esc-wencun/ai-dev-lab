// 密码校验规则（个人中心改密）—— 对位基准 infoPwdValidator（pwdChrtype 0-4 动态）

import { pwdValidatorFactory } from '@/utils/passwordRule'

export const infoPwdValidator = pwdValidatorFactory('info')
