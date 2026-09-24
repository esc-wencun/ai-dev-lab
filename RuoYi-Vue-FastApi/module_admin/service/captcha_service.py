import io
import os
import random
import base64
from PIL import Image, ImageDraw, ImageFont


class CaptchaService:
    """
    验证码模块服务层（数字计算型验证码，与java版captchaType=math一致）
    """

    @classmethod
    async def create_captcha_image_service(cls):
        """
        生成算术验证码图片
        :return: [base64图片字符串, 计算结果]
        """
        # 创建空白图像
        image = Image.new('RGB', (160, 60), color='#EAEAEA')

        # 创建绘图对象
        draw = ImageDraw.Draw(image)

        # 设置字体（优先使用项目自带字体，找不到时使用PIL默认字体）
        font = None
        font_path = os.path.join(os.path.abspath(os.getcwd()), 'assets', 'font', 'Arial.ttf')
        if os.path.exists(font_path):
            font = ImageFont.truetype(font_path, size=30)
        else:
            try:
                # windows系统字体
                font = ImageFont.truetype('arial.ttf', size=30)
            except Exception:
                font = ImageFont.load_default()

        # 生成两个随机整数
        num1 = random.randint(0, 9)
        num2 = random.randint(0, 9)
        # 从运算符列表中随机选择一个
        operational_character_list = ['+', '-', 'x']
        operational_character = random.choice(operational_character_list)
        # 根据选择的运算符进行计算
        if operational_character == '+':
            result = num1 + num2
        elif operational_character == '-':
            result = num1 - num2
        else:
            result = num1 * num2

        # 绘制文本
        text = f"{num1} {operational_character} {num2} = ?"
        # 随机文字颜色
        color = (random.randint(0, 150), random.randint(0, 150), random.randint(0, 150))
        draw.text((25, 15), text, fill=color, font=font)

        # 绘制干扰线
        for _ in range(6):
            x1 = random.randint(0, 160)
            y1 = random.randint(0, 60)
            x2 = random.randint(0, 160)
            y2 = random.randint(0, 60)
            draw.line([(x1, y1), (x2, y2)], fill=(random.randint(0, 200), random.randint(0, 200), random.randint(0, 200)))

        # 将图像数据保存到内存中
        buffer = io.BytesIO()
        image.save(buffer, format='JPEG')

        # 将图像数据转换为base64字符串
        base64_string = base64.b64encode(buffer.getvalue()).decode()

        return [base64_string, result]
