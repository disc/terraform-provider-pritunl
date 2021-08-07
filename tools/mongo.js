use pritunl;
db.administrators.update(
    {
        "username": "pritunl"
    },
    {
        $set: {
            auth_api: true,
            token: "tfacctest_token",
            secret: "tfacctest_secret"
        }
    }
)