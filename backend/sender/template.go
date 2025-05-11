package sender

const emailTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
        }
        .header {
            background-color: #A855F7;
            color: white;
            padding: 20px;
            text-align: center;
            border-radius: 8px 8px 0 0;
        }
        .content {
            background-color: #ffffff;
            padding: 20px;
            border: 1px solid #F3E8FF;
            border-radius: 0 0 8px 8px;
        }
        .footer {
            text-align: center;
            margin-top: 20px;
            color: #6b7280;
            font-size: 12px;
        }
        .button {
            display: inline-block;
            background-color: #A855F7;
            color: white !important;
            padding: 10px 20px;
            text-decoration: none;
            border-radius: 4px;
            margin: 20px 0;
        }
        .button:hover {
            background-color: #7E22CE;
        }
        .logo {
            width: 60px;
            height: 60px;
            margin: 10px auto;
            display: block;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>{{.EventName}}</h1>
    </div>
    <div class="content">
        <p>Hey {{.Username}}, your photos from {{.EventName}} are ready to download. Click the button below to download them:</p>
        <div style="text-align: center;">
            <a href="{{.DownloadLink}}" class="button">Download Photos</a>
        </div>
        <p>If the button doesn't work, you can right-click on the button above and select "Copy Link Address" to paste in your browser.</p>
        <p>This link will expire in 7 days.</p>
    </div>
    <div class="footer">
        <img src="https://files.arsln.dev/fd-logo-purple.png" alt="FaceDrop Logo" class="logo">
        <p>© 2024 FaceDrop. All rights reserved.</p>
        <p>This is an automated message, please do not reply to this email.</p>
    </div>
</body>
</html>
` 