const getJson = (res) => res.json();

function getHeaders() {
  const headers = new Headers();
  headers.append("Authorization", 'Bearer ' + window.localStorage.getItem('token'));

  return headers;
}

function post(url, body) {
  const opts = {
    body: JSON.stringify(body),
    headers: getHeaders(),
    method: 'POST'
  };

  return fetch(url, opts).then(getJson);
}

function postForm(url, data) {
  const opts = {
    body: data,
    headers: getHeaders(),
    method: 'POST'
  };

  return fetch(url, opts).then(getJson);
}

function get(url) {
  const opts = {
    headers: getHeaders(),
    method: 'GET'
  };

  return fetch(url, opts).then(getJson);
}

function delete_(url) {
  const opts = {
    headers: getHeaders(),
    method: 'DELETE'
  };

  return fetch(url, opts).then(getJson);
}

function put(url, body) {
  const opts = {
    body: JSON.stringify(body),
    headers: getHeaders(),
    method: 'PUT'
  };

  return fetch(url, opts).then(getJson);
}

function getAll(selector) {
  return Array.prototype.slice.call(document.querySelectorAll(selector), 0);
}

$(document).ready(function() {
  const $burgers = getAll('.burger');

 if ($burgers.length > 0) {
   $burgers.forEach($el => {
     $el.addEventListener('click', () => {
       const target = $el.dataset.target;
       const $target = document.getElementById(target);
       $el.classList.toggle('is-active');
       $target.classList.toggle('is-active');
     });
   });
 }
});