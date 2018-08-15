var bus = new Vue();

// COMMONS //

var BOOKS = 'books', AUTHORS = 'authors', SERIES = 'series';
var BOUQUINS_TYPES = {
  books: { icon: 'book', singular: 'livre', plural: 'livres',
    tab_cols:  [ { id: 'title',   name: 'Titre', sort: 'title' },
                 { id: 'authors', name: 'Auteur(s)' },
                 { id: 'series',  name: 'Serie' } ] },
  authors: { icon: 'user', singular: 'auteur', plural: 'auteurs',
    tab_cols: [ { id: 'author_name', name: 'Nom', sort: 'name' },
                { id: 'count',       name: 'Livre(s)' } ] },
  series: { icon: 'list', singular: 'serie', plural: 'series',
    tab_cols: [ { id: 'serie_name', name: 'Nom', sort: 'name' },
                { id: 'count',      name: 'Livre(s)' },
                { id: 'authors',    name: 'Auteur(s)' } ] }
};
function ty(type) {
  if (BOUQUINS_TYPES[type]) return BOUQUINS_TYPES[type]
  console.log("ERROR: Unknown type: " + type);
  return {}
}
function icon(type) {
  return ty(type).icon;
}
function iconClass(type) {
  return 'glyphicon glyphicon-' + icon(type);
}
function url(type, id) {
  if (id) return ty(type) ? '/'+type+'/'+id:'';
  return ty(type) ? '/'+type+'/':'';
}
function label(type, count) {
  return count == 1 ? ty(type).singular : ty(type).plural;
}
function stdError(code, resp) {
  console.log('ERROR ' + code + ': ' + resp);
}
function sendQuery(url, error, success) {
  var xmh = new XMLHttpRequest();
  var v;
  xmh.onreadystatechange = function() {
    v = xmh.responseText;
    if (xmh.readyState === 4 && xmh.status === 200) {
      var res;
      try  { 
        res = JSON.parse(v);
      } catch (err) {
        if (null !== error)
          error(err.name, err.message);
      }
      if (null !== success)
        success(res);
    } else if (xmh.readyState === 4) {
      if (null !== error)
        error(xmh.status, v);
    }
  };
  xmh.open('GET', url, true);
  xmh.setRequestHeader('Accept','application/json');
  xmh.send(null);
}

// COMPONENTS //

Vue.component('results-list', {
  template: '#results-list-template',
  props: ['results', 'count', 'type'],
  methods: {
    url: function(item) { return url(this.type, item.id); },
    label: function(item) {
      switch (this.type) {
        case BOOKS:
          return item.title;
        case AUTHORS:
        case SERIES:
          return item.name;
        default:
          return '';
      }
    },
    iconClass: function() {
      return iconClass(this.type);
    },
    countlabel: function() {
      return label(this.type, this.count);
    }
  }
});
Vue.component('results', {
  template: '#results-template',
  props: ['results', 'cols','sort_by','order_desc'],
  methods: {
    sortBy: function(col) {
      bus.$emit('sort-on', col);
    }
  }
});
Vue.component('result-cell', {
  render: function(h) {
    return h('td', this.cellContent(h));
  },
  props: ['item', 'col'],
  methods: {
    link: function(h, type, text, id) {
      return [
        h('span',{ attrs: { class: iconClass(type) } },''),
        ' ',
        h('a', { attrs: { href: url(type, id) } }, text)
      ];
    },
    badge: function(h, num) {
      return h('span', { attrs: { class: 'badge' } }, num);
    },
    cellContent: function(h) {
      switch (this.col.id) {
      case 'author_name':
        return this.link(h, AUTHORS, this.item.name, this.item.id);
      case 'serie_name':
        return this.link(h, SERIES, this.item.name, this.item.id);
      case 'count':
        return this.item.count;
      case 'title':
        return this.link(h, BOOKS, this.item.title, this.item.id);
      case 'authors':
        var elts = [];
        var authors = this.item.authors;
        if (authors) {
          for (i=0;i<authors.length;i++) {
            elts[i] = this.link(h, AUTHORS, authors[i].name, authors[i].id);
          }
        }
        return elts;
      case 'series':
        var series = this.item.series;
        if (series) {
          return [
            this.link(h, SERIES, series.name, series.id),
            h('span', { attrs: { class: 'badge' } }, this.item.series_idx)
          ];
        }
        return '';
      default:
        console.log('ERROR unknown col: ' + this.col.id)
        return '';
      }
    }
  }
});
Vue.component('paginate', {
  template: '#paginate-template',
  props: ['page','more'],
  methods: {
    prevPage: function() {
      if (this.page > 1) bus.$emit('update-page', -1);
    },
    nextPage: function() {
      if (this.more) bus.$emit('update-page', 1);
    }
  }
});

// PAGES //

if (document.getElementById("index")) {
  new Vue({
    el: '#index',
    data: {
      url: '',
      page: 0,
      perpage: 20,
      more: false,
      sort_by: null,
      order_desc: false,
      cols: [],
      results: []
    },
    methods: {
      sortBy: function(col) {
        if (this.sort_by == col) {
          if (this.order_desc) {
            this.order_desc = false;
            this.sort_by = null;
          } else {
            this.order_desc = true;
          }
        } else {
          this.order_desc = false;
          this.sort_by = col;
        }
        this.updateResults();
      },
      updatePage: function(p) {
        this.page += p;
        this.updateResults();
      },
      order: function(query) {
        return query + (this.order_desc ? '&order=desc' : '');
      },
      sort: function(query) {
        return query + (this.sort_by ? '&sort=' + this.sort_by : '');
      },
      paginate: function(query) {
        return query + '?page=' + this.page + '&perpage=' + this.perpage;
      },
      params: function(url) {
        return this.order(this.sort(this.paginate(url)));
      },
      updateResults: function() {
        sendQuery(this.params(this.url), stdError, this.loadResults);
      },
      showSeries: function() {
        this.url = url(SERIES);
        this.updateResults();
      },
      showAuthors: function() {
        this.url = url(AUTHORS);
        this.updateResults();
      },
      showBooks: function() {
        this.url = url(BOOKS);
        this.updateResults();
      },
      loadCols: function(type) {
        this.cols = ty(type).tab_cols;
      },
      loadResults(resp) {
        this.results = [];
        this.more = resp.more;
        this.loadCols(resp.type);
        if (resp.results) {
          this.results = resp.results;
          if (this.page == 0) this.page = 1;
        } else {
          this.page = 0;
        }
      }
    },
    mounted: function() {
      bus.$on('sort-on', this.sortBy);
      bus.$on('update-page', this.updatePage);
    }
  });
}
if (document.getElementById("author")) {
  new Vue({
    el: '#author',
    data: {
      tab: BOOKS
    },
    methods: {
      showBooks: function() {
        this.tab = BOOKS;
      },
      showAuthors: function() {
        this.tab = AUTHORS;
      },
      showSeries: function() {
        this.tab = SERIES;
      }
    }
  });
}

if (document.getElementById("search")) {
  new Vue({
    el: '#search',
    data: {
      urlParams: [],
      authors: [],
      books: [],
      series: [],
      authorsCount: 0,
      booksCount: 0,
      seriesCount: 0,
      q: '',
      which: 'all',
      all: false,
      perpage: 10
    },
    methods: {
      searchParams: function(url) {
        var res = url + '?perpage=' + this.perpage;
        for (var i=0; i<this.terms.length; i++) {
          var t = this.terms[i];
          if (t.trim())
            res += '&term=' + encodeURIComponent(t.trim());
        }
        return res;
      },
      searchAuthorsSuccess: function(res) {
        this.authorsCount = res.count;
        this.authors = res.results;
      },
      searchAuthors: function() {
        sendQuery(this.searchParams(url(AUTHORS)), stdError, this.searchAuthorsSuccess);
      },
      searchBooksSuccess: function(res) {
        this.booksCount = res.count;
        this.books = res.results;
      },
      searchBooks: function() {
        sendQuery(this.searchParams(url(BOOKS)), stdError, this.searchBooksSuccess);
      },
      searchSeriesSuccess: function(res) {
        this.seriesCount = res.count;
        this.series = res.results;
      },
      searchSeries: function() {
        sendQuery(this.searchParams(url(SERIES)), stdError, this.searchSeriesSuccess);
      },
      searchAll: function() {
        this.searchAuthors();
        this.searchBooks();
        this.searchSeries();
      },
      clear: function() {
        this.authors = [];
        this.books = [];
        this.series = [];
        this.authorsCount = 0;
        this.booksCount = 0;
        this.seriesCount = 0;
      },
      searchFull: function() {
        if (this.q) {
          this.terms = this.q.split(' ');
          this.clear();
          switch (this.which) {
            case AUTHORS:
              this.searchAuthors();
              break;
            case BOOKS:
              this.searchBooks();
              break;
            case SERIES:
              this.searchSeries();
              break;
            default:
              this.searchAll();
              break;
          }
        }
        return false;
      },
      searchUrl: function() {
        if (this.urlParams.q) {
          this.terms = this.urlParams.q.split(' ');
          this.clear();
          this.searchAll();
          this.q = this.urlParams.q;
        }
      },
      urlParse: function() {
        var match,
        pl     = /\+/g,  // Regex for replacing addition symbol with a space
        search = /([^&=]+)=?([^&]*)/g,
        decode = function (s) { return decodeURIComponent(s.replace(pl, " ")); },
        query  = window.location.search.substring(1);
        while (match = search.exec(query))
          this.urlParams[decode(match[1])] = decode(match[2]);
      }  
    },
    created: function() {
      this.urlParse();
    },
    mounted: function() {
      this.searchUrl();
    }
  });
}
