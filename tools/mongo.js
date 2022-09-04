use pritunl;
db.administrators.updateOne(
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