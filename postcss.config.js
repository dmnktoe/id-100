const fs = require('fs');
const path = require('path');

// Store for CSS Modules mappings
const cssModules = {};

module.exports = {
  plugins: {
    'postcss-import': {
      path: ['src', 'web/static'],
      resolve: (id, basedir, importOptions) => {
        // Handle node_modules imports like 'swiper/css'
        if (!id.startsWith('.') && !id.startsWith('/')) {
          try {
            return require.resolve(id, { paths: [basedir] });
          } catch (e) {
            // Fallback to default resolution
          }
        }
        return id;
      }
    },
    'postcss-modules': {
      // Generate scoped class names with hash
      generateScopedName: '[name]__[local]___[hash:base64:5]',
      // Callback to collect CSS module mappings
      getJSON: (cssFileName, json, outputFileName) => {
        // Merge all class mappings
        Object.assign(cssModules, json);
        
        // Write to file when processing is done
        const outputPath = path.join(__dirname, 'web/static/css-modules.json');
        fs.writeFileSync(outputPath, JSON.stringify(cssModules, null, 2));
      }
    },
    cssnano: {
      preset: ['default', {
        discardComments: {
          removeAll: true
        }
      }]
    }
  }
};
