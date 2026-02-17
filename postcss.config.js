module.exports = {
  plugins: {
    "postcss-import": {
      path: ["src", "web/static"],
      resolve: (id, basedir, importOptions) => {
        // Handle node_modules imports like 'swiper/css'
        if (!id.startsWith(".") && !id.startsWith("/")) {
          try {
            return require.resolve(id, { paths: [basedir] });
          } catch (e) {
            // Fallback to default resolution
          }
        }
        return id;
      },
    },
    cssnano: {
      preset: [
        "default",
        {
          discardComments: {
            removeAll: true,
          },
        },
      ],
    },
  },
};
